// node.go

package main

import (
	"log"

	cmn "github.com/nsip/n3-transport/common"
	"github.com/nsip/n3-transport/n3node"
)

//
// Launches a dispatcher to manage user
// connections and messaging on n3
//
func main() {

	n3c, err := n3node.NewNode()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("node running...")

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		n3c.Close()
	})

}
