// streams.go

package n3liftbridge

import (
	"context"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
)

//
// calls liftbridge to create a stream
// subject - the nats topic the stream will connect to
// name - the name of the stream
//
func CreateStream(serverAddr, subject, name string) error {

	// connect to liftbridge server
	lbClient, err := liftbridge.Connect([]string{serverAddr})
	if err != nil {
		return err
	}

	// define the stream
	newStream := liftbridge.StreamInfo{
		Subject:           subject,
		Name:              name,
		ReplicationFactor: 1,
	}

	if err := lbClient.CreateStream(context.Background(), newStream); err != nil {
		if err != nil && err != liftbridge.ErrStreamExists {
			return err
		}
	}

	return nil

}
