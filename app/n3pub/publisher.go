// publisher.go

package main

import (
	"log"

	"github.com/nats-io/nuid"
	cmn "github.com/nsip/n3-transport/common"
	"github.com/nsip/n3-transport/messages"
	"github.com/nsip/n3-transport/n3liftbridge"
)

func main() {

	n3pub, err := n3liftbridge.NewPublisher()
	if err != nil {
		log.Fatalln("Fatal error starting publisher: ", err)
	}

	log.Println("publisher running, Ctrl+C to exit...")

	// send some messages
	for i := 0; i < 10; i++ {

		subject := nuid.Next()
		predicate := nuid.Next()
		object := nuid.Next()
		context := "TR"

		t, err := messages.NewTuple(subject, predicate, object, context)
		if err != nil {
			log.Fatal("cannot create new tuple: ", err)
		}

		// assign arbitrary version - done by CMS in real system
		t.Version = 1

		err = n3pub.Publish(t)
		if err != nil {
			log.Fatal(err)
		}

	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		err = n3pub.Close()
		if err != nil {
			log.Println("error closing: ", err)
		}
		log.Println("publisher successfully shut down")

	})

}
