// n3validator.go

//
// launches the abci app
//

package main

import (
	"log"
	"os"

	"github.com/nsip/n3-transport/n3tendermint"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/abci/types"

	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

func main() {
	err := initN3Validator()
	if err != nil {
		log.Fatal(err)
	}
}

func initN3Validator() error {
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))

	// Create the application
	var app types.Application

	app = n3tendermint.NewN3Application()

	// Start the listener
	srv, err := server.NewServer("tcp://127.0.0.1:26658", "socket", app)
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		return err
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
	return nil
}
