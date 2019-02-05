// n3publisher.go

package main

import (
	"flag"
	"log"

	"github.com/nsip/n3-messages/messages"
	"github.com/nsip/n3-messages/n3grpc"
)

//
// test harness example of creating a publisher to talk to the
// grpc endpoint on an n3 node
//
//
func main() {

	numMessages := flag.Int("n", 1000, "number of messages to publish")
	nameSpace := flag.String("namespace", "", "namespace id for the context to publish to")
	contextName := flag.String("context", "", "context name to publish to")

	flag.Parse()

	if *contextName == "" || *nameSpace == "" {
		log.Fatalln("must provide namespace and context name")
	}

	n3pub, err := n3grpc.NewPublisher("localhost", 5777)
	if err != nil {
		log.Fatalln("cannot create publisher:", err)
	}
	defer n3pub.Close()

	tuple, err := messages.NewTuple("subject1", "predicate1longlonglonglonglonglonglonglong", "obj1")

	for i := 0; i < *numMessages; i++ {
		tuple.Version = int64(i)
		err = n3pub.Publish(tuple, *nameSpace, *contextName)
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
