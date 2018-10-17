// publisher.go

package n3grpc

import (
	"context"
	"log"

	"github.com/nsip/n3-transport/messages"
	"github.com/nsip/n3-transport/messages/pb"
	"github.com/pkg/errors"
)

//
// publisher is the business-level wrapper around the
// grpc client, that exposes domain-friendly methods
// for sending new tuples/messages to an n3 node
//
type Publisher struct {
	grpcClient pb.APIClient
	stream     pb.API_PublishClient
}

func NewPublisher(host string, port int) (*Publisher, error) {

	client, err := newAPIClient(host, port)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create grpc client")
	}
	stream, err := client.Publish(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create stream connection to server")
	}
	pub := &Publisher{grpcClient: client, stream: stream}

	return pub, nil
}

//
// constructs a trasport-level message from the tup,e and the delivery params (namespace, contextName)
// and sends to the grpc server on the n3 node
//
// tuple - an SPO Tuple (no version required, will be assinged by node)
// namespace - owner of the delivery context - a base58 pub key id
// contextName - the delivery context for this data
//
func (pub *Publisher) Publish(tuple *pb.SPOTuple, namespace, contextName string) error {

	// encode the tuple
	payload, err := messages.EncodeTuple(tuple)
	if err != nil {
		return err
	}

	// set the message envelope parameters & payload
	n3msg := &pb.N3Message{
		Payload:   payload,
		NameSpace: namespace,
		CtxName:   contextName,
	}

	// send the message
	err = pub.stream.Send(n3msg)
	if err != nil {
		return err
	}

	return nil
}

//
// closes the network connections to the grpc server, and retrieves
// a tx summary report of how many messages have been sent from this
// publisher.
//
func (pub *Publisher) Close() {

	// get the tx report
	reply, err := pub.stream.CloseAndRecv()
	if err != nil {
		log.Printf("%v.CloseAndRecv() got error %v, want %v", pub.stream, err, nil)
	}
	log.Printf("Tx Summary: %v", reply)

}
