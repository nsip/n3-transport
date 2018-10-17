// messages.go

package messages

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/nsip/n3-transport/messages/pb"
	"github.com/pkg/errors"
)

// functions for constructing tuples and rendering as messages

//
// create a new tuple, validating all fields have been provided
//
func NewTuple(subject, predicate, object string) (*pb.SPOTuple, error) {

	if subject == "" {
		return nil, errors.New("subject must be provided")
	}

	if predicate == "" {
		return nil, errors.New("predicate must be provided")
	}

	// nil objects are allowed to 'remove' existing values

	tuple := &pb.SPOTuple{
		Subject:   strings.TrimSpace(subject),
		Predicate: strings.TrimSpace(predicate),
		Object:    strings.TrimSpace(object),
	}

	return tuple, nil
}

//
// returns a protobuf encoded message for transmission
//
func NewMessage(payload []byte, sender, namespace, contextName string) ([]byte, error) {

	msg := &pb.N3Message{
		Payload:   payload,
		SndId:     sender,
		NameSpace: namespace,
		CtxName:   contextName,
	}

	pbMsg, err := EncodeN3Message(msg)
	if err != nil {
		return nil, err
	}

	return pbMsg, nil
}

//
// retrieve n3message from protobuf transmission format
//
func DecodeN3Message(message []byte) (*pb.N3Message, error) {

	n3msg := &pb.N3Message{}
	if err := proto.Unmarshal(message, n3msg); err != nil {
		return nil, errors.Wrap(err, "unable to decode n3 msg from protbuf message:")
	}

	return n3msg, nil

}

//
// returns a protobuf encoded tuple for transmission / encryption
//
func EncodeN3Message(m *pb.N3Message) ([]byte, error) {

	n3msg, err := proto.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "protobuf encoding failed for message:")
	}
	return n3msg, nil
}

//
// from a given protobuf message returns the tuple
//
func DecodeTuple(message []byte) (*pb.SPOTuple, error) {

	tuple := &pb.SPOTuple{}
	if err := proto.Unmarshal(message, tuple); err != nil {
		return nil, errors.Wrap(err, "unable to decode tuple from protbuf message:")
	}

	return tuple, nil

}

//
// returns a protobuf encoded tuple for transmission / encryption
//
func EncodeTuple(t *pb.SPOTuple) ([]byte, error) {

	pbTuple, err := proto.Marshal(t)
	if err != nil {
		return nil, errors.Wrap(err, "protobuf encoding failed for tuple:")
	}
	return pbTuple, nil
}
