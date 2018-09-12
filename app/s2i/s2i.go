// s2i.go

package main

import (
	"flag"
	"log"

	cmn "github.com/nsip/n3-transport/common"
	inf "github.com/nsip/n3-transport/n3influx"
)

//
// simple application that reads from a given context stream and
// writes the data into influx
//
func main() {

	contextPtr := flag.String("context", "SIF", "name of context to connect to")
	flag.Parse()

	ic, err := inf.NewInfluxConnector(*contextPtr)
	if err != nil {
		log.Fatal(err)
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		err = ic.Close()
		if err != nil {
			log.Println("error closing: ", err)
		}
		log.Println("s2i successfully shut down")

	})

}
