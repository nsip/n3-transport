// dispatcher.go

package main

import (
	"log"

	cmn "github.com/nsip/n3-client/common"
	"github.com/nsip/n3-client/n3dispatcher"
)

//
// Launches a dispatcher to manage user
// connections and messaging on n3
//
func main() {

	n3disp, err := n3dispatcher.NewDispatcher()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("dispatcher running...")

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		n3disp.Close()
	})

}
