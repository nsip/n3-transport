// client.go

package n3grpc

import (
	"fmt"

	"github.com/nsip/n3-transport/messages/pb"
	"google.golang.org/grpc"
)

//
// creates a grpc client that connects to the specified server hostname:port
// for practical use, is embedded in the publisher component.
//
func newAPIClient(host string, port int) (pb.APIClient, error) {

	serverAddr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure()) // TODO: upgrade to non-insecure if required ie. jwt/tls etc.
	if err != nil {
		return nil, err
	}
	return pb.NewAPIClient(conn), nil

}
