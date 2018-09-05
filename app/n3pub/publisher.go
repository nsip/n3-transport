// publisher.go

package main

import (
	"log"

	cmn "github.com/nsip/n3-transport/common"
	"github.com/nsip/n3-transport/n3liftbridge"
)

func main() {

	n3pub, err := n3liftbridge.NewPublisher()
	if err != nil {
		log.Fatalln("Fatal error starting publisher: ", err)
	}

	log.Println("publisher running, Ctrl+C to exit...")

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
