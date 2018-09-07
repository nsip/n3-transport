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
func NewTuple(subject, predicate, object, context string) (*pb.SPOTuple, error) {

	if subject == "" {
		return nil, errors.New("subject must be provided")
	}

	if predicate == "" {
		return nil, errors.New("predicate must be provided")
	}

	// nil objects are allowed to 'remove' existing values

	if context == "" {
		return nil, errors.New("context must be provided")
	}

	tuple := &pb.SPOTuple{
		Subject:   strings.TrimSpace(subject),
		Predicate: strings.TrimSpace(predicate),
		Object:    strings.TrimSpace(object),
		Context:   strings.TrimSpace(context),
	}

	return tuple, nil
}

//
// returns a protobuf encoded tuple for transmission
//
func NewMessage(t *pb.SPOTuple) ([]byte, error) {

	pbTuple, err := proto.Marshal(t)
	if err != nil {
		return nil, errors.Wrap(err, "protobuf encoding failed for tuple:")
	}
	return pbTuple, nil
}

//
// returns a specialised tuple used for requesting access to
// context for a user.
//
// user - requesting access for this user, if blank will be current system user
//
//
func NewTrustRequest(user, topic string) (*pb.SPOTuple, error) {

}

//
// from a given protobuf message returns the tuple
//
func Decode(message []byte) (*pb.SPOTuple, error) {

	tuple := &pb.SPOTuple{}
	if err := proto.Unmarshal(message, tuple); err != nil {
		return nil, errors.Wrap(err, "unable to decode tuple from protbuf message:")
	}

	return tuple, nil

}
