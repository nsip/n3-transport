// n3liftbridge_publisher.go

package n3liftbridge

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
	lbproto "github.com/liftbridge-io/go-liftbridge/liftbridge-grpc"
	nats "github.com/nats-io/go-nats"
	"github.com/nsip/n3-transport/pb"
	"github.com/pkg/errors"
)

type Publisher struct {
	TRStream       liftbridge.StreamInfo            // stream definition for Trust Requests
	TRContext      context.Context                  // conext to manage stream
	trCtxCancel    func()                           // cancel func for context
	trustRequests  []*pb.SPOTuple                   // slice of known trust requests
	TAStream       liftbridge.StreamInfo            // stream definition for Trust Approvals
	TAContext      context.Context                  // context to manage stream
	taCtxCancel    func()                           // cancel func for context
	natsConn       *nats.Conn                       // connection to nats server for msg publishing
	lbClient       liftbridge.Client                // connect to liftbridge streams
	StreamRegistry map[string]liftbridge.StreamInfo // register of approved user streams
	registryMutex  sync.Mutex                       // mutex to protect registry in multi-threads
}

func NewPublisher() (*Publisher, error) {

	trstream := createTRStreamInfo()
	tastream := createTAStreamInfo()
	log.Println("stream infos created")

	// connection to nats server
	natsconn, err := nats.GetDefaultOptions().Connect()
	if err != nil {
		return nil, errors.Wrap(err, "Nats connection failed")
	}
	log.Println("nats connection established")

	// connection to liftbridge
	addr := "localhost:9292"
	lbclient, err := liftbridge.Connect([]string{addr})
	if err != nil {
		return nil, errors.Wrap(err, "Liftbridge connection failed")
	}
	log.Println("liftbridge connection established")

	// initialise the stream registry
	registry := make(map[string]liftbridge.StreamInfo)

	// set up contexts for managing liftbridge interactions
	trCtx, trCancelFunc := context.WithCancel(context.Background())
	taCtx, taCancelFunc := context.WithCancel(context.Background())

	pub := &Publisher{
		TRStream:       trstream,
		TRContext:      trCtx,
		trCtxCancel:    trCancelFunc,
		trustRequests:  make([]*pb.SPOTuple, 0),
		TAStream:       tastream,
		TAContext:      taCtx,
		taCtxCancel:    taCancelFunc,
		natsConn:       natsconn,
		lbClient:       lbclient,
		StreamRegistry: registry,
	}
	log.Println("candidate publisher created")

	// ensure auth streams are set up
	err = pub.verifyTrustStreams()
	if err != nil {
		return nil, err
	}
	log.Println("trust streams verified")

	// start handlers for trust messages
	log.Println("starting trust handlers")
	pub.startTRHandler()
	pub.startTAHandler()

	log.Println("publisher available...")

	return pub, nil

}

//
// subscribes to the trust request channel
//
func (pub *Publisher) startTRHandler() {

	// logic invoked when a message is received
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("error in TR handler: ", err)
			// return
		}
		log.Println(msg.Offset, string(msg.Value))
		// create stream info based on TA
		// add to registry
	}

	// create the subscription and run until context is cancelled
	go func() {
		subscriptionError := pub.lbClient.Subscribe(pub.TRContext, pub.TRStream.Subject,
			pub.TRStream.Name, handler)
		if subscriptionError != nil {
			pub.trCtxCancel()
		}

		<-pub.TRContext.Done()
		log.Println("TR context cancelled, handler closing; last error: ", subscriptionError)

	}()

	log.Println("...TR handler up")
}

//
// subscribes to the trust approval channel
//
func (pub *Publisher) startTAHandler() {

	// logic invoked when a message is received
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("error in TA handler: ", err)
			// return
		}
		log.Println(msg.Offset, string(msg.Value))
		// create stream info based on TA
		// add to registry
	}

	// create the subscription and run until context is cancelled
	go func() {
		subscriptionError := pub.lbClient.Subscribe(pub.TAContext, pub.TAStream.Subject,
			pub.TAStream.Name, handler)
		if subscriptionError != nil {
			pub.taCtxCancel()
		}

		<-pub.TAContext.Done()
		log.Println("TA context cancelled, handler closing; last error: ", subscriptionError)

	}()

	log.Println("...TA handler up")

}

//
// Publisher must have access to trust request/approval streams, so
// always force creation
//
func (pub *Publisher) verifyTrustStreams() error {

	// create TR stream
	err := pub.lbClient.CreateStream(pub.TRContext, pub.TRStream)
	if err != liftbridge.ErrStreamExists && err != nil {
		return err
	}
	log.Println("TR stream created")

	// create TA stream
	err = pub.lbClient.CreateStream(pub.TAContext, pub.TAStream)
	if err != liftbridge.ErrStreamExists && err != nil {
		return err
	}
	log.Println("TA stream created")

	return nil
}

//
// ensure all connections cleanly closed.
//
func (pub *Publisher) Close() error {

	log.Println("closing trust contexts")
	pub.taCtxCancel()
	pub.trCtxCancel()

	log.Println("closing liftbridge connection")
	err := pub.lbClient.Close()
	if err != nil {
		return err
	}

	log.Println("closing nats connection")
	err = pub.natsConn.Flush()
	if err != nil {
		return err
	}
	pub.natsConn.Close()

	log.Println("waiting for close...")
	time.Sleep(time.Second * 2)

	return nil

}

//
// creates the stream definition for the Trust Request stream
//
func createTRStreamInfo() liftbridge.StreamInfo {

	return liftbridge.StreamInfo{
		Subject:           "TR",
		Name:              "TR-Stream",
		ReplicationFactor: 1,
	}
}

//
// creates the stream definition for the Traust Approval stream
//
func createTAStreamInfo() liftbridge.StreamInfo {

	return liftbridge.StreamInfo{
		Subject:           "TA",
		Name:              "TA-Stream",
		ReplicationFactor: 1,
	}

}

//
//
//
func createStream(userid string) error {

	if userid == "" {
		return errors.New("cannot have zero-length userid")
	}

	addr := "localhost:9292"
	client, err := liftbridge.Connect([]string{addr})
	if err != nil {
		return err
	}
	defer client.Close()
	stream := liftbridge.StreamInfo{
		Subject:           userid,
		Name:              userid + "-stream",
		ReplicationFactor: 1,
	}
	if err := client.CreateStream(context.Background(), stream); err != nil {
		if err != liftbridge.ErrStreamExists {
			return err
		}
	}
	fmt.Println("created stream ", stream.Name)
	return nil
}
