// strm2influx.go

package n3influx

import (
	"context"
	"fmt"
	"log"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
	lbproto "github.com/liftbridge-io/go-liftbridge/liftbridge-grpc"
	"github.com/nsip/n3-transport/messages"
	"github.com/pkg/errors"
)

//
// Reads tuples from a liftbridge stream and
// writes to an influx db
//
type InfluxConnector struct {
	lbClient    liftbridge.Client // connect to liftbridge streams
	infPub      *Publisher        // influx publisher
	contextName string            // name of the context stream to connect
}

//
// creates a new connector that reads from the given stream neame
// and stores the read tuples into influx
//
// contextName is the aggregate stream to be read, which will be translated
// into the following Name and Subject parameters for the stream subscription:
//
// cotextName: MyModel
// subscription Suject -> MyModel.*
// subscritpion Name -> MyModel.stream
//
func NewInfluxConnector(contextname string) (*InfluxConnector, error) {

	if contextname == "" {
		return nil, errors.New("context name must be provided")
	}

	// connection to liftbridge
	addr := "localhost:9292"
	lbclient, err := liftbridge.Connect([]string{addr})
	if err != nil {
		return nil, errors.Wrap(err, "Liftbridge connection failed")
	}
	log.Println("liftbridge connection established")

	infpub, err := NewPublisher()
	if err != nil {
		return nil, errors.Wrap(err, "Influx connection failed")
	}
	log.Println("influx connection established")

	ic := &InfluxConnector{
		lbClient:    lbclient,
		infPub:      infpub,
		contextName: contextname,
	}

	ic.startStreamHandler()

	return ic, nil
}

//
// starts stream reader and publishes tuples to influx
//
func (iv *InfluxConnector) startStreamHandler() {

	aggregateSubject := fmt.Sprintf("%s.*", iv.contextName)
	aggregateStreamName := fmt.Sprintf("%s.stream", iv.contextName)

	// logic invoked when a message is received
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("error in stream handler: ", err)
			return
		}

		// decode from stream
		streamTuple, err := messages.Decode(msg.Value)
		if err != nil {
			log.Println(err)
			return
		}

		// publish to influx
		err = iv.infPub.StoreTuple(streamTuple)
		if err != nil {
			log.Println("unable to publish tuple to influx: ", err)
			return
		}

	}

	// create the subscription and run until context is cancelled
	go func() {
		ctx, ctxCancel := context.WithCancel(context.Background())
		subscriptionError := iv.lbClient.Subscribe(ctx, aggregateSubject,
			aggregateStreamName, handler)
		if subscriptionError != nil {
			ctxCancel()
		}

		<-ctx.Done()
		log.Println("StreamHandler context cancelled, handler closing; last error: ", subscriptionError)

	}()

	log.Println("...Stream handler up")

}
