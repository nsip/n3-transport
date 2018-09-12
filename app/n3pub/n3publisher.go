// n3publisher.go

package main

import (
	"log"
	"time"

	"github.com/nats-io/nuid"
	n3transport "github.com/nsip/n3-transport"
	cmn "github.com/nsip/n3-transport/common"
	"github.com/nsip/n3-transport/messages"
)

func main() {

	n3pub, err := n3transport.NewPublisher()
	if err != nil {
		log.Fatalln("Fatal error starting publisher: ", err)
	}

	log.Println("publisher running, Ctrl+C to exit...")

	// create a trust request
	topic := "SIF"   // name of model context
	user := "NSIP01" // should be pk id of user who wants access
	tr, err := messages.NewTrustRequest(user, topic)
	if err != nil {
		log.Fatal("cannot create trust request: ", err)
	}
	// _ = tr
	// publish trust request
	err = n3pub.PublishToStream(tr)
	if err != nil {
		log.Fatal("cannot publish TR: ", err)
	}

	// allow time to work through the log
	log.Println("processing Trust Requests")
	time.Sleep(time.Second * 5)

	// get current trust requests, and approve one
	trs, err := n3pub.TrustRequests()
	if err != nil {
		log.Fatal(err)
	}
	if len(trs) == 0 {
		log.Println("no outstanding TRs to process")
	}

	// for testing take the first one
	// and convert to a trust approval
	ta, err := messages.NewTrustApproval(trs[0])
	if err != nil {
		log.Fatal("cannot create TA: ", err)
	}
	// _ = ta
	// publish the trust approval
	err = n3pub.PublishToStream(ta)
	if err != nil {
		log.Fatal(err)
	}

	// allow time to work through the log
	log.Println("processing Trust Approvals")
	time.Sleep(time.Second * 5)

	// // send some messages
	log.Println("sending messages...")

	for i := 0; i < 300; i++ {

		subject := nuid.Next()
		predicate := nuid.Next()
		object := nuid.Next()
		context := "SIF"

		t, err := messages.NewTuple(subject, predicate, object, context)
		if err != nil {
			log.Fatal("cannot create new tuple: ", err)
		}

		// assign arbitrary version - done by CMS in real system
		t.Version = 1

		// err = n3pub.PublishToBlockchain(t)
		err = n3pub.PublishToStream(t)
		if err != nil {
			log.Fatal(err)
		}
		// log.Println("message published")

	}
	log.Println("...all messages published")

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
