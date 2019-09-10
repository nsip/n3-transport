// dispatcher.go

package n3dispatcher

import (
	"context"
	"fmt"
	"log"
	"sync"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
	lbproto "github.com/liftbridge-io/liftbridge-grpc/go"
	nats "github.com/nats-io/go-nats"
	"github.com/nsip/n3-messages/messages"
	"github.com/nsip/n3-transport/n3config"
	"github.com/nsip/n3-transport/n3liftbridge"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//
// Dispatcher listens for user.context approvals stream
// and creates handlers to process received messages
// to then forward to the user feed
//
type Dispatcher struct {
	natsConn        *nats.Conn
	lbConn          liftbridge.Client
	handlerContexts map[string]func()
	pubKey          string
	privKey         string
	mutex           sync.Mutex
}

func NewDispatcher() (*Dispatcher, error) {

	initConfig()

	// get the b58 string of the public key
	b58pubkey := viper.GetString("pubkey")
	log.Println("this dispatcher id is: ", b58pubkey)
	b58privkey := viper.GetString("privkey")

	// create a nats client connection
	natsurl := viper.GetString("nats_addr")
	natsconn, err := nats.Connect(natsurl)
	if err != nil {
		return nil, errors.Wrap(err, "dispatcher cannot connect to Nats")
	}

	// create a liftbridge client connection
	lbaddr := viper.GetString("lb_addr")
	lbconn, err := liftbridge.Connect([]string{lbaddr})
	if err != nil {
		return nil, errors.Wrap(err, "dispatcher cannot connect to Liftbridge")
	}

	// ensure approvals stream is available
	err = n3liftbridge.CreateStream(lbaddr, "approvals", "approvals-stream")
	if err != nil {
		return nil, errors.Wrap(err, "dispatcher cannot connect to approvals stream")
	}
	// ensure messages stream is available
	err = n3liftbridge.CreateStream(lbaddr, b58pubkey, b58pubkey+"-stream")
	if err != nil {
		return nil, errors.Wrap(err, "dispatcher cannot connect to messages stream")
	}

	// initialise the handler contexts map
	handlercontexts := make(map[string]func())

	disp := &Dispatcher{
		pubKey:          b58pubkey,
		privKey:         b58privkey,
		natsConn:        natsconn,
		lbConn:          lbconn,
		handlerContexts: handlercontexts,
	}

	err = disp.startJoinHandler()
	if err != nil {
		return nil, err
	}

	err = disp.startApprovalHandler()
	if err != nil {
		return nil, err
	}

	return disp, nil

}

//
// reads a confiduration file if exists
// if not then creates a default config & saves
//
func initConfig() {

	err := n3config.ReadConfig()
	if err == n3config.ConfigNotFoundError {
		err = n3config.CreateBaseConfig()
		if err != nil {
			log.Fatalln("cannot create dipatcher config:", err)
		}
	} else if err != nil {
		log.Fatalln("cannot initialise config:", err)
	}

	log.Println("Using config file:", viper.ConfigFileUsed())
}

//
// creates a req/response handler that will give dispatcher info to
// a client joining the network
//
func (disp *Dispatcher) startJoinHandler() error {

	// create context to control async subscription
	ctx, cancelFunc := context.WithCancel(context.Background())
	disp.addHandlerContext("join", cancelFunc)

	// handle request messages - when user joins reply with
	// public key of this dispatcher
	handler := func(msg *nats.Msg) {
		log.Println("join request accepted from: ", string(msg.Data))
		err := disp.natsConn.Publish(msg.Reply, []byte(disp.pubKey))
		if err != nil {
			log.Println("join response error: ", err)
			return
		}
	}

	// start up the listener
	go func() {
		_, err := disp.natsConn.Subscribe("join", handler)
		if err != nil {
			log.Println("erorr in join handler: ", err)
			disp.removeHandlerContext("join")
		}
		<-ctx.Done()
	}()

	return nil

}

//
// reads from the approvals stream and starts a worker for each
// user.context connection
//
func (disp *Dispatcher) startApprovalHandler() error {

	// create context to control async subscription
	ctx, cancelFunc := context.WithCancel(context.Background())
	disp.addHandlerContext("approvals", cancelFunc)

	// logic invoked when an approval message is received
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("approval handler received error from liftbridge: ", err)
			return
		}

		// decode n3 message from transmission protobuf format
		// approvals are not point-to-point encrypted as they need
		// to be visible to all dispatchers
		n3msg, err := messages.DecodeN3Message(msg.Value)
		if err != nil {
			log.Println("cannot decode message proto: ", err)
			return
		}
		log.Println("received approval from:", n3msg.SndId)

		// decode the internal tuple
		approvalTuple, err := messages.DecodeTuple(n3msg.Payload)
		if err != nil {
			log.Println("cannot decode tuple proto: ", err)
			return
		}

		// start or stop worker
		workerUser := approvalTuple.Subject
		workerScope := fmt.Sprintf("%s.%s", n3msg.NameSpace, approvalTuple.Object)

		if approvalTuple.Predicate == "grant" {
			err = disp.createWorker(workerUser, workerScope)
			if err != nil {
				log.Println("unable to create worker: ", err)
				return
			}
		}

		if approvalTuple.Predicate == "deny" {
			handlerTag := fmt.Sprintf("%s.%s", workerUser, workerScope)
			disp.removeHandlerContext(handlerTag)
			return
		}
	}

	// create the subscription and run until context is cancelled
	go func() {
		err := disp.lbConn.Subscribe(ctx, "approvals-stream", handler, liftbridge.StartAtEarliestReceived())
		if err != nil {
			log.Println("error subscribing approvals handler: ", err)
			disp.removeHandlerContext("approvals")
		}
		<-ctx.Done()
	}()

	return nil
}

//
// invokes a new handler that will process and forward
// messages to this user.context scope
//
func (disp *Dispatcher) createWorker(user, scope string) error {

	// create user.owner.contextName worker, if not already up!
	handlerTag := fmt.Sprintf("%s.%s", user, scope)
	if !disp.haveHandler(handlerTag) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		disp.addHandlerContext(handlerTag, cancelFunc)
		cfg := &workerConfig{
			ctx:        ctx,
			user:       user,
			scope:      scope,
			handlerTag: handlerTag,
			disp:       disp,
		}
		err := startWorker(cfg)
		if err != nil {
			disp.removeHandlerContext(handlerTag)
			return err
		}
		log.Println("worker started for: ", user, scope)
		return nil
	}
	log.Println("handler already running for: ", user, scope)

	return nil
}

//
// reports whether a handler is already running for a
// given topic
//
func (disp *Dispatcher) haveHandler(name string) bool {

	disp.mutex.Lock()
	defer disp.mutex.Unlock()

	_, ok := disp.handlerContexts[name]

	return ok
}

//
// adds the cancel function for context used to control the given handler
//
func (disp *Dispatcher) addHandlerContext(name string, cancelFunc func()) {

	disp.mutex.Lock()
	defer disp.mutex.Unlock()

	disp.handlerContexts[name] = cancelFunc

}

//
// invokes the given handler context cancel function, and removes the
// handler from the internal list
//
func (disp *Dispatcher) removeHandlerContext(name string) {

	disp.mutex.Lock()
	defer disp.mutex.Unlock()

	cfunc, ok := disp.handlerContexts[name]
	if ok {
		cfunc()
		delete(disp.handlerContexts, name)
		log.Println("handler shut down: ", name)
	}

}

//
// shuts down connections, closes handlers
//
func (disp *Dispatcher) Close() {

	// shut down nats connection
	disp.natsConn.Flush()
	disp.natsConn.Close()

	// shut down lb connection
	disp.lbConn.Close()

	// close all handlers by invoking cancelfunc on associated contexts
	for name, _ := range disp.handlerContexts {
		disp.removeHandlerContext(name)
	}

	log.Println("dispatcher successfully shut down")

}
