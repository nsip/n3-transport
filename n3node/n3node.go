// n3node.go

package n3node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"../n3config"
	"../n3crypto"
	"../n3influx"
	"../n3liftbridge"
	"../n3nats"
	"github.com/google/uuid"
	liftbridge "github.com/liftbridge-io/go-liftbridge"
	lbproto "github.com/liftbridge-io/go-liftbridge/liftbridge-grpc"
	nats "github.com/nats-io/go-nats"
	"github.com/nsip/n3-messages/messages"
	"github.com/nsip/n3-messages/messages/pb"
	"github.com/nsip/n3-messages/n3grpc"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	u "github.com/cdutwhu/go-util"
)

//
// main node tht handles read / write operations
// with the transport layer
//
type N3Node struct {
	natsConn         *nats.Conn
	lbConn           liftbridge.Client
	handlerContexts  map[string]func()
	dispatcherId     string
	mutex            sync.Mutex
	pubKey           string
	privKey          string
	approvedContexts map[string]bool
}

func NewNode() (*N3Node, error) {

	err := checkConfig()
	if err != nil {
		return nil, errors.Wrap(err, "ensure you have run 'n3cli init'")
	}

	// get the b58 string of the public key
	b58pubkey := viper.GetString("pubkey")
	log.Println("this node id is: ", b58pubkey)
	b58privkey := viper.GetString("privkey")

	// create a nats client connection
	natsurl := viper.GetString("nats_addr")
	natsconn, err := nats.Connect(natsurl)
	if err != nil {
		return nil, errors.Wrap(err, "node cannot connect to Nats")
	}

	// create a liftbridge client connection
	lbaddr := viper.GetString("lb_addr")
	lbconn, err := liftbridge.Connect([]string{lbaddr})
	if err != nil {
		return nil, errors.Wrap(err, "node cannot connect to Liftbridge")
	}

	// initialise the handler contexts map
	handlercontexts := make(map[string]func())

	// initilaise the approved contexts map
	approvedContexts := make(map[string]bool)

	n3c := &N3Node{
		pubKey:           b58pubkey,
		privKey:          b58privkey,
		natsConn:         natsconn,
		lbConn:           lbconn,
		handlerContexts:  handlercontexts,
		approvedContexts: approvedContexts,
	}

	// ensure node read stream is available
	streamName := fmt.Sprintf("%s-stream", n3c.pubKey)
	err = n3liftbridge.CreateStream(lbaddr, n3c.pubKey, streamName)
	if err != nil {
		return nil, errors.Wrap(err, "node cannot connect to read stream")
	}

	//
	// listen for approvals and allow this node
	// to send to approved contexts
	//
	err = n3c.startApprovalHandler()
	if err != nil {
		n3c.Close()
		return nil, err
	}

	//
	// write handler listens for messages on the grpc server
	// for this client and sends to dispatcher
	//
	err = n3c.startWriteHandler()
	if err != nil {
		n3c.Close()
		return nil, err
	}

	//
	// reads all messages from this client's feed
	// and passes on to storage handlers
	//
	err = n3c.startReadHandler()
	if err != nil {
		return nil, err
	}

	return n3c, nil

}

//
// checks that a config file exists
//
func checkConfig() error {

	return n3config.ReadConfig()
}

//
// reads from the approvals stream and records the
// approval state for target streams this client can
// send messages to
//
func (n3c *N3Node) startApprovalHandler() error {

	//
	// logic invoked when an approval message is received
	//
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("node approval handler received error from liftbridge: ", err)
			return
		}

		// decode n3 message from transmission protobuf format
		// approvals are not point-to-point encrypted as they need
		// to be visible to all dispatchers and clients
		n3msg, err := messages.DecodeN3Message(msg.Value)
		if err != nil {
			log.Println("node cannot decode message proto: ", err)
			return
		}
		log.Println("node received approval from:", n3msg.SndId)

		// decode the internal tuple
		approvalTuple, err := messages.DecodeTuple(n3msg.Payload)
		if err != nil {
			log.Println("node cannot decode approval tuple proto: ", err)
			return
		}

		// record approval
		approvalScope := fmt.Sprintf("%s.%s.%s", approvalTuple.Subject, n3msg.NameSpace, approvalTuple.Object)
		log.Println("approval scope: ", approvalScope)
		n3c.mutex.Lock()
		defer n3c.mutex.Unlock()

		// handle the approval
		if approvalTuple.Predicate == "grant" {
			n3c.approvedContexts[approvalScope] = true
		}
		if approvalTuple.Predicate == "deny" {
			n3c.approvedContexts[approvalScope] = false
		}
	}

	// create context to control async subscription
	ctx, cancelFunc := context.WithCancel(context.Background())
	n3c.addHandlerContext("approvals", cancelFunc)

	// create the subscription and run until context is cancelled
	go func() {
		err := n3c.lbConn.Subscribe(ctx, "approvals", "approvals-stream", handler, liftbridge.StartAtEarliestReceived())
		if err != nil {
			log.Println("node error subscribing approvals handler: ", err)
			n3c.removeHandlerContext("approvals")
		}
		<-ctx.Done()
	}()

	return nil
}

//
// listens for new messages from this client and sends to
// dispatcher
//
func (n3c *N3Node) startWriteHandler() error {

	// get an instance of direct db query
	dbClient, _ := n3influx.NewPublisher()

	// join network, acquire dispatcher
	dispatcherid, err := n3nats.Join(n3c.natsConn, n3c.pubKey)
	if err != nil {
		return errors.Wrap(err, "node closing: no dispatcher available.")
	}

	// set up handler for inbound messages
	handler := func(n3msg *pb.N3Message) {

		// force sender id to be this node
		n3msg.SndId = n3c.pubKey

		// unpack tuple payload
		tuple, err := messages.DecodeTuple(n3msg.Payload)
		if err != nil {
			log.Println("write handler cannot decode tuple: ", err)
			return
		}

		// check authorisation
		approvalScope := fmt.Sprintf("%s.%s.%s", n3msg.SndId, n3msg.NameSpace, n3msg.CtxName)
		if !n3c.approved(approvalScope) {
			log.Println("you are not authorised to send to context: ", n3msg.NameSpace, n3msg.CtxName)
			return
		}

		// TODO: check privacy rules

		// TODO: assign lamport clock version
		if !assignVer(dbClient, tuple, n3msg.CtxName) {
			return
		}

		// specify dispatcher
		// n3msg.DispId = dispatcherid

		// encrypt tuple payload for dispatcher
		encryptedTuple, err := n3crypto.EncryptTuple(tuple, dispatcherid, n3c.privKey)
		if err != nil {
			log.Println("write handler unable to encrypt tuple: ", err)
			return
		}

		newMsg := &pb.N3Message{
			Payload:   encryptedTuple,
			SndId:     n3c.pubKey,
			NameSpace: n3msg.NameSpace,
			CtxName:   n3msg.CtxName,
			DispId:    dispatcherid,
		}

		// encode & send
		msgBytes, err := messages.EncodeN3Message(newMsg)
		if err != nil {
			log.Println("write handler unable to encode message: ", err)
			return
		}

		err = n3c.natsConn.Publish(dispatcherid, msgBytes)
		if err != nil {
			log.Println("write handler unable to publish message: ", err)
			return
		}
	}

	// *** set up handler for query by inbound messages ***
	qHandler := func(n3msg *pb.N3Message) (ts []*pb.SPOTuple) {
		fPln("in Query qHandler")
		tuple := Must(messages.DecodeTuple(n3msg.Payload)).(*pb.SPOTuple)
		s := tuple.Subject
		start, end, _ := getValueVerRange(dbClient, s, n3msg.CtxName)

		if sHS(n3msg.CtxName, "-sif") {

			tempCtx := fSf("temp_%d", time.Now().UnixNano())
			n := dbClient.BatTrans(tuple, n3msg.CtxName, tempCtx, false, true, start, end)
			fPln("v", n)

			tupleS := &pb.SPOTuple{Subject: tuple.Predicate, Predicate: "::"}
			n = dbClient.BatTrans(tupleS, n3msg.CtxName, tempCtx, true, false, 0, 0)
			fPln("s", n)

			tupleA := &pb.SPOTuple{Subject: tuple.Predicate, Predicate: tuple.Subject}
			n = dbClient.BatTrans(tupleA, n3msg.CtxName, tempCtx, true, false, 0, 0)
			fPln("a", n)

			/******************************************/
			arrInfo := make(map[string]int) /* key: predicate, value: array count */
			// ctx, revArr := n3msg.CtxName, true /* Search from original measurement, reverse array order */
			// dbClient.QueryTuple(tuple, 0, revArr, ctx, &ts, &arrInfo, start, end) /* Search from original measurement */
			ctx, revArr := tempCtx, true                                    /* Search from temp measurement, reverse array order */
			dbClient.QueryTuple(tuple, 0, revArr, ctx, &ts, &arrInfo, 0, 0) /* Search from temp measurement */
			dbClient.AdjustOptionalTuples(&ts, &arrInfo)                    /* we need re-order some tuples */

			/******************************************/
			dbClient.DropCtx(tempCtx)

		} else if sHS(n3msg.CtxName, "-xapi") {
			dbClient.QueryTuples(tuple, n3msg.CtxName, &ts, start, end)
		} else if sHS(n3msg.CtxName, "-meta") { // *** request a ticket for publishing

		AGAIN:
			if _, ok := mapTickets.Load(s); ok {
				time.Sleep(time.Millisecond * DELAY_CONTEST)
				goto AGAIN
			}

			_, ve, v := getValueVerRange(dbClient, tuple.Subject, n3msg.CtxName)
			termID, endV := uuid.New().String(), u.I64(ve).ToStr()
			ts = append(ts, &pb.SPOTuple{Subject: s, Predicate: termID, Object: endV, Version: v}) // *** return result ***

			mapTickets.Store(s, &ticket{tktID: termID, idx: endV})

			if flagRmTicket {
				go ticketRmAsync(dbClient, &mapTickets, n3msg.CtxName)
				flagRmTicket = false
			}
		}
		return
	}

	// start server
	apiServer := n3grpc.NewAPIServer()
	apiServer.SetMessageHandler(handler, qHandler)
	return apiServer.Start(viper.GetInt("rpc_port"))
}

//
// checks if sending to the given context is allowed for this user
//
func (n3c *N3Node) approved(scope string) bool {
	n3c.mutex.Lock()
	defer n3c.mutex.Unlock()

	approved, ok := n3c.approvedContexts[scope]
	if !ok {
		return false
	}

	return approved

}

//
// reads from this client's feed and passes to
// storage / query layer
//
func (n3c *N3Node) startReadHandler() error {

	// start up the influx publisher
	// TODO: abstract multi-db handlers to channel reader
	// to multiplex send to different data-stores
	pub, err := n3influx.NewPublisher()
	if err != nil {
		return errors.Wrap(err, "read handler cannot connect to influx store:")
	}

	// set up message handler
	lastMessageOffset := viper.GetInt64(n3c.pubKey)
	var nextMessage int64
	if lastMessageOffset != 0 {
		nextMessage = lastMessageOffset + 1
	}
	log.Println("consming messages offset from:", nextMessage)
	handler := func(msg *lbproto.Message, err error) {

		if err != nil {
			log.Println("read handler error from liftbridge server:", err)
			return
		}

		// decode msg from transport format
		n3msg, err := messages.DecodeN3Message(msg.Value)
		if err != nil {
			log.Println("node read handler cannot decode message proto: ", err)
			return
		}
		// log.Println("node received message from:", n3msg.SndId)

		// unpack payload - decrypt and unmarshal
		tuple, err := n3crypto.DecryptTuple(n3msg.Payload, n3msg.DispId, n3c.privKey)
		if err != nil {
			log.Println("read handler decrypt error: ", err)
			return
		}

		// *** exclude "legend liftbridge data" ***
		if inDB(pub, tuple, n3msg.CtxName) {
			return
		}

		err = pub.StoreTuple(tuple, n3msg.CtxName)
		if err != nil {
			log.Println("error storing tuple:", err)
			return
		}
		// log.Println("...tuple stored successfully.")
		lastMessageOffset = msg.Offset

	}

	// create context to control async subscription
	ctx, cancelFunc := context.WithCancel(context.Background())
	n3c.addHandlerContext(n3c.pubKey, cancelFunc)

	// create the subscription and run until context is cancelled
	go func() {
		streamName := fmt.Sprintf("%s-stream", n3c.pubKey)
		err := n3c.lbConn.Subscribe(ctx, n3c.pubKey, streamName, handler, liftbridge.StartAtOffset(nextMessage))
		if err != nil {
			log.Println("node error subscribing read handler: ", err)
			n3c.removeHandlerContext(n3c.pubKey)
		}
		<-ctx.Done()
		// store the last read position
		viper.Set(n3c.pubKey, lastMessageOffset)
	}()

	return nil
}

//
// adds the cancel function for context used to control the given handler
//
func (n3c *N3Node) addHandlerContext(name string, cancelFunc func()) {

	n3c.mutex.Lock()
	defer n3c.mutex.Unlock()

	n3c.handlerContexts[name] = cancelFunc

}

//
// invokes the given handler context cancel function, and removes the
// handler from the internal list
//
func (n3c *N3Node) removeHandlerContext(name string) {

	n3c.mutex.Lock()
	defer n3c.mutex.Unlock()

	cfunc, ok := n3c.handlerContexts[name]
	if ok {
		cfunc()
		delete(n3c.handlerContexts, name)
		log.Println("node context handler shut down: ", name)
	}

}

//
// shuts down connections, closes handlers
//
func (n3c *N3Node) Close() {

	// shut down nats connection
	n3c.natsConn.Flush()
	n3c.natsConn.Close()

	// shut down lb connection
	n3c.lbConn.Close()

	// close all handlers by invoking cancelfunc on associated contexts
	for name, _ := range n3c.handlerContexts {
		n3c.removeHandlerContext(name)
	}

	// save the config file, to remember client read position
	err := n3config.SaveConfig()
	if err != nil {
		log.Println("unable to save config:", err)
	}

	log.Println("node successfully shut down")

}
