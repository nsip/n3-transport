// n3publisher.go

package main

import (
	"log"

	"github.com/nsip/n3-transport/messages"
	"github.com/nsip/n3-transport/n3grpc"
)

//
// test harness example of creating a publisher to talk to the
// grpc endpoint on an n3 node
//
func main() {

	n3pub, err := n3grpc.NewPublisher("localhost", 5777)
	if err != nil {
		log.Fatalln("cannot create publisher:", err)
	}
	defer n3pub.Close()

	namespace := "EjonwQe6SD2BrWeMz69q4RNnZxAZFmGsA3cNAXTLHdPz" // from config in real world
	contextName := "mfcontext"

	tuple, err := messages.NewTuple("subject1", "predicate1longlonglonglonglonglonglonglong", "obj1")

	for i := 0; i < 5; i++ {
		tuple.Version = int64(i)
		err = n3pub.Publish(tuple, namespace, contextName)
		if err != nil {
			log.Fatalln("publish error:", err)
		}
	}

	log.Println("messages sent")

	// // Wait forever
	// cmn.TrapSignal(func() {
	// 	// Cleanup
	// 	n3pub.Close()
	// 	log.Println("publisher successfully shut down")

	// })

}
