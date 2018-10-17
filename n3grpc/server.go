// server.go

package n3grpc

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/nsip/n3-transport/messages/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

//
// implementation of the grpc server used by the node
// to receive 'raw' messages from outside the n3
// environment
// API is defined in messages/pb/n3msg.proto
//

type MessageHandler func(*pb.N3Message)

type APIServer struct {
	msgHandler MessageHandler
	grpcServer *grpc.Server
}

//
// creates the API server which implements the Publish method
// and wraps in a generic grpc server which can be launched
// with normal tcp dial params elsewhere - e.g. in the
// context of the n3 node, from within the client
//
// the server will simply consume (nullify) messages unless a
// handler is provided to supply some business logic.
//
func NewAPIServer() *APIServer {

	apiServer := &APIServer{
		grpcServer: grpc.NewServer(),
		msgHandler: func(msg *pb.N3Message) { msg = nil },
	}
	// bind this api server logic into the generic grpc server
	pb.RegisterAPIServer(apiServer.grpcServer, apiServer)

	return apiServer

}

//
// creates a running api server on the given port
//
func (s *APIServer) Start(port int) error {

	// get underlying tcp connection
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.Wrap(err, "cannot start grpc server: ")
	}

	// launch the server
	go func() {
		log.Println("grpc server running on port", port)
		s.grpcServer.Serve(lis)
	}()

	return nil
}

//
// the server simply receives n3messages in a stream from clients.
// the handler allows business logic to be added by the creator of
// server to process messages
//
func (s *APIServer) SetMessageHandler(mh MessageHandler) {
	s.msgHandler = mh
}

func (s *APIServer) Publish(stream pb.API_PublishServer) error {

	var msgCount int64

	for {
		n3msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.TxSummary{
				MsgCount: msgCount,
			})
		}
		if err != nil {
			return err
		}
		// do something with the msg
		s.msgHandler(n3msg)
		msgCount++

	}
}
