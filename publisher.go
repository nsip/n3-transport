// publisher.go

package n3transport

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
	lbproto "github.com/liftbridge-io/go-liftbridge/liftbridge-grpc"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/nuid"
	"github.com/nsip/n3-transport/messages"
	"github.com/nsip/n3-transport/messages/pb"
	tm "github.com/nsip/n3-transport/n3tendermint"
	"github.com/pkg/errors"
)

type Publisher struct {
	TRStream       liftbridge.StreamInfo            // stream definition for Trust Requests
	TRContext      context.Context                  // conext to manage stream
	trCtxCancel    func()                           // cancel func for context
	trustRegister  map[string]*pb.SPOTuple          // slice of known trust requests
	TAStream       liftbridge.StreamInfo            // stream definition for Trust Approvals
	TAContext      context.Context                  // context to manage stream
	taCtxCancel    func()                           // cancel func for context
	natsConn       *nats.Conn                       // connection to nats server for msg publishing
	lbClient       liftbridge.Client                // connect to liftbridge streams
	streamRegister map[string]liftbridge.StreamInfo // register of approved user streams
	registerMutex  sync.Mutex                       // mutex to protect registers in multi-threads
	tmPub          *tm.Publisher                    // tendermint publisher client
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

	// connection to tendermint b/c
	tmpub, err := tm.NewPublisher()
	if err != nil {
		return nil, errors.Wrap(err, "Tendermint connection failed")
	}
	log.Println("tendermint connection established")

	// initialise trust request registry
	tr_register := make(map[string]*pb.SPOTuple)
	// initialise the trust approval stream registry
	str_register := make(map[string]liftbridge.StreamInfo)

	// set up contexts for managing liftbridge interactions
	trCtx, trCancelFunc := context.WithCancel(context.Background())
	taCtx, taCancelFunc := context.WithCancel(context.Background())

	pub := &Publisher{
		TRStream:       trstream,
		TRContext:      trCtx,
		trCtxCancel:    trCancelFunc,
		trustRegister:  tr_register,
		TAStream:       tastream,
		TAContext:      taCtx,
		taCtxCancel:    taCancelFunc,
		natsConn:       natsconn,
		lbClient:       lbclient,
		streamRegister: str_register,
		tmPub:          tmpub,
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
// returns the current list of un-porcessed trust requests
//
func (pub *Publisher) TrustRequests() ([]*pb.SPOTuple, error) {

	trs := make([]*pb.SPOTuple, 0)
	pub.registerMutex.Lock()
	for _, tr := range pub.trustRegister {
		trs = append(trs, tr)
	}
	pub.registerMutex.Unlock()

	return trs, nil

}

//
// Publish messages to the required stream
// Takes a new SPO Tuple and adds user details, signature etc.
// Validates user for the chosen stream using Check() internally
// Publishes to the stream using Deliver()
//
func (pub *Publisher) PublishToStream(t *pb.SPOTuple) error {

	// connect data to this user
	err := pub.authoriseTuple(t)
	if err != nil {
		return errors.Wrap(err, "unable to authorise tuple for transmssion")
	}

	// validation checks
	err = pub.Check(t)
	if err != nil {
		return err
	}

	// publish
	err = pub.Deliver(t)
	if err != nil {
		return err
	}

	return nil

}

//
// Publish message to the blockchain transport which will then
// distribute to all other b/c nodes.
// Takes a new SPO Tuple and adds user details, signature etc.
// Only the initial Check() validation is run, as decision to Deliver()
// is made by the blockchain validators themselves
//
func (pub *Publisher) PublishToBlockchain(t *pb.SPOTuple) error {

	// connect data to this user
	err := pub.authoriseTuple(t)
	if err != nil {
		return errors.Wrap(err, "unable to authorise tuple for transmssion")
	}

	// validation checks
	err = pub.Check(t)
	if err != nil {
		return err
	}

	// publish
	err = pub.bcDeliver(t)
	if err != nil {
		return err
	}

	return nil

}

//
// Check, runs basic validation on a supplied tuple
// checks signature against content and validates that this
// user is allowed to publish to this stream (context)
//
func (pub *Publisher) Check(t *pb.SPOTuple) error {

	// verify signed msg - TODO
	// check whether stream is autorised
	if !pub.authorisedStream(t) {
		return errors.New("not authorised to publish to context: " + t.Context)
	}

	return nil

}

//
// Deliver, publishes the message to the streaming service
// to the stream specified by tuple.Context
//
func (pub *Publisher) Deliver(t *pb.SPOTuple) error {

	tname := pub.getTargetTopicName(t)

	// proto encode the message
	msg, err := messages.NewMessage(t)
	if err != nil {
		return err
	}

	// messages are published to nats, not stream
	if err := pub.natsConn.Publish(tname, msg); err != nil {
		return errors.Wrap(err, "could not publish msg to nats")
	}

	return nil

}

//
// bcDeliver, publishes the message to the blockchain transport
// unexported method as should only be called via PublishToBlockchain
//
func (pub *Publisher) bcDeliver(t *pb.SPOTuple) error {

	// proto encode the message
	msg, err := messages.NewMessage(t)
	if err != nil {
		return err
	}

	err = pub.tmPub.SubmitTx(msg)
	if err != nil {
		return errors.Wrap(err, "could not pubish to tendermint")
	}

	return nil

}

//
// given a data tuple add required user meta-data and signature
//
func (pub *Publisher) authoriseTuple(t *pb.SPOTuple) error {

	t.User = "NSIP01"       // to be replaced with public key
	t.Signature = "sig0001" // to be replaced with pk sig of message
	t.MsgID = nuid.Next()
	t.Version = 0 // should come from CMS when integrated

	return nil

}

//
// determines whether the user has access to the stream
//
func (pub *Publisher) authorisedStream(t *pb.SPOTuple) bool {
	switch {
	case t.Context == "TA":
		return true
	case t.Context == "TR":
		return true
	default:
		userKey := pub.getTargetTopicName(t)
		// log.Println("userkey: ", userKey)
		_, ok := pub.streamRegister[userKey]
		return ok
	}
}

//
// builds a nats topic name from the contents of the tuple
// if tuple.Context is TA/TR these are the the topic names returned
// otherwise topic is [tuple.Context].[tuple.User]
//
func (pub *Publisher) getTargetTopicName(t *pb.SPOTuple) string {

	if t.Context == "TA" || t.Context == "TR" {
		return t.Context
	}

	return fmt.Sprintf("%s.%s", t.Context, t.User)

}

//
// subscribes to the trust request channel
//
func (pub *Publisher) startTRHandler() {

	// logic invoked when a message is received
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("error in TR handler: ", err)
			return
		}

		tuple, err := messages.Decode(msg.Value)
		if err != nil {
			log.Println(err)
			return
		}

		// add the trust request to the registry
		pub.registerMutex.Lock()
		pub.trustRegister[tuple.MsgID] = tuple
		pub.registerMutex.Unlock()

		// log.Println("TR: ", msg.Offset, fmt.Sprintf("%v", tuple))

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
			return
		}

		taTuple, err := messages.Decode(msg.Value)
		if err != nil {
			log.Println(err)
			return
		}

		// log.Println("TA: ", msg.Offset, fmt.Sprintf("%v", taTuple))

		// check TA against approvl rules
		if !pub.meetsApprovalRules(taTuple) {
			log.Println("TR not yet approved, does not meet rules.")
			return
		}
		// create approved streams based on TA
		err = pub.createApprovedStreams(taTuple)
		if err != nil {
			log.Println("unable to create approved streams: ", err)
			return
		}
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
// enforces approval rules for access to streams
//
func (pub *Publisher) meetsApprovalRules(ta *pb.SPOTuple) bool {
	return true // obviously change this!!!
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
// from the TA tuple, ensure that the relevant streams
// exist.
//
func (pub *Publisher) createApprovedStreams(ta *pb.SPOTuple) error {

	req, ok := pub.trustRegister[ta.Subject]
	if !ok {
		return errors.New("no request found for TA")
	}

	userSubject := fmt.Sprintf("%s.%s", ta.Object, req.Subject)
	userStreamName := fmt.Sprintf("%s.%s-stream", ta.Object, req.Subject)
	aggregateSubject := fmt.Sprintf("%s.*", ta.Object)
	aggregateStreamName := fmt.Sprintf("%s.stream", ta.Object)

	// see if already created
	pub.registerMutex.Lock()
	_, ok = pub.streamRegister[userSubject]
	pub.registerMutex.Unlock()
	if ok {
		return nil
	}

	// create streaminfo for desired stream/user combination
	userStream := liftbridge.StreamInfo{
		Subject:           userSubject,
		Name:              userStreamName,
		ReplicationFactor: 1,
	}
	// create stream info for aggregate stream
	aggregateStream := liftbridge.StreamInfo{
		Subject:           aggregateSubject,
		Name:              aggregateStreamName,
		ReplicationFactor: 1,
	}

	// create streams on broker, ignore errors if alsready created
	if err := pub.lbClient.CreateStream(context.Background(), userStream); err != nil {
		if err != liftbridge.ErrStreamExists {
			return err
		}
	}
	log.Println("created user stream - ", userStream.Name)

	if err := pub.lbClient.CreateStream(context.Background(), aggregateStream); err != nil {
		if err != liftbridge.ErrStreamExists {
			return err
		}
	}
	log.Println("created aggregate stream - ", aggregateStream.Name)

	// store infos in stream registry as approved routes -- for user!!!
	pub.registerMutex.Lock()
	pub.streamRegister[userSubject] = userStream
	pub.registerMutex.Unlock()

	log.Println("approved streams created & registered")
	return nil
}
