// node.go

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	cmn "github.com/nsip/n3-transport/common"
	"github.com/nsip/n3-transport/n3node"
	"github.com/pkg/errors"
)

var cnclGnatsd, cnclLiftbridge, cnclDispatcher, cnclInflux func()
var launchErr error
var n3n n3node.N3Node

//
// Launches a dispatcher to manage user
// connections and messaging on n3
//
func main() {

	runServices := flag.Bool("withServices", true, "launches all background services (gnats, liftbridge, influx, dispatcher) set to false if only running as a consumer node.")
	servicesOnly := flag.Bool("servicesOnly", false, "launches just background services (gnats, liftbridge, influx, dispatcher).")

	flag.Parse()

	if *runServices {
		launchErr = launchServices()
		if launchErr != nil {
			closeServices()
			log.Fatalln("unable to start services:", launchErr)
		}
	}

	if !*servicesOnly {
		n3n, err := n3node.NewNode()
		if err != nil {
			if *runServices {
				closeServices()
			}
			log.Fatalln("unable to create node:", err)
		}
		defer n3n.Close()
		log.Println("n3node running...")
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		if *runServices {
			closeServices()
			log.Println("...all services shut down")
		}
	})

}

func launchServices() error {

	var serviceError error
	cnclGnatsd, serviceError = launchGnatsd()
	if serviceError != nil {
		return errors.Wrap(serviceError, "cannot launch gnatsd service:")
	}
	log.Println("...gnats service up.")
	cnclLiftbridge, serviceError = launchLiftbridge()
	if serviceError != nil {
		return errors.Wrap(serviceError, "cannot launch liftbridge service:")
	}
	log.Println("...liftbridge service up.")
	// always launch a dispatcher so there's at least one
	// on the network
	cnclDispatcher, serviceError = launchDispatcher()
	if serviceError != nil {
		return errors.Wrap(serviceError, "cannot launch dispatcher service:")
	}
	log.Println("...disptcher service up.")
	cnclInflux, serviceError = launchInflux()
	if serviceError != nil {
		return errors.Wrap(serviceError, "cannot launch influx service:")
	}
	log.Println("...influx service up.")

	return nil
}

func launchInflux() (func(), error) {

	// influx logging is verbose, so send to a local file
	logFile, err := os.Create("./services/influx/influxlog.txt")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create influx log file")
	}

	// set runtime arguments
	config := "--config=influxdb.conf"
	args := []string{config}

	// set up a management context
	ctx, cancel := context.WithCancel(context.Background())

	influxCmd := exec.CommandContext(ctx, "./influxd", args...)
	influxCmd.Stdout = os.Stdout
	influxCmd.Stderr = logFile
	influxCmd.Dir = "./services/influx"
	err = influxCmd.Start()
	if err != nil {
		return nil, err
	}

	// give the server time to come up
	time.Sleep(time.Second * 2)

	return cancel, nil

}

func launchDispatcher() (func(), error) {

	// set up a management context
	ctx, cancel := context.WithCancel(context.Background())

	dispCmd := exec.CommandContext(ctx, "./n3dispatcher")
	dispCmd.Stdout = os.Stdout
	dispCmd.Stderr = os.Stderr
	dispCmd.Dir = "./n3dispatcher"
	err := dispCmd.Start()
	if err != nil {
		return nil, err
	}

	// give the server time to come up
	time.Sleep(time.Second * 2)

	return cancel, nil

}

func launchGnatsd() (func(), error) {

	// set up a management context
	ctx, cancel := context.WithCancel(context.Background())

	gnatsCmd := exec.CommandContext(ctx, "./gnatsd")
	gnatsCmd.Stdout = os.Stdout
	gnatsCmd.Stderr = os.Stderr
	gnatsCmd.Dir = "./services/gnatsd"
	err := gnatsCmd.Start()
	if err != nil {
		return nil, err
	}

	// give the server time to come up
	time.Sleep(time.Second * 2)

	return cancel, nil

}

func launchLiftbridge() (func(), error) {

	// set runtime arguments
	raft := "--raft-bootstrap-seed"
	dir := "--data-dir=./services/liftbridge/data"
	args := []string{raft, dir}

	// set up a management context
	ctx, cancel := context.WithCancel(context.Background())

	lbCmd := exec.CommandContext(ctx, "./liftbridge", args...)
	lbCmd.Stdout = os.Stdout
	lbCmd.Stderr = os.Stderr
	lbCmd.Dir = "./services/liftbridge"
	err := lbCmd.Start()
	if err != nil {
		return nil, err
	}

	// give the server time to come up
	time.Sleep(time.Second * 3)

	return cancel, nil

}

func closeServices() {
	cnclDispatcher()
	log.Println("...dispatcher service shut down")
	cnclLiftbridge()
	log.Println("...liftbridge service shut down")
	cnclGnatsd()
	log.Println("...gnatsd service shut down")
	cnclInflux()
	log.Println("...influx service shut down")

	// need time for cancels to propagate otherwise get
	// indeterminate shutdowns
	time.Sleep(time.Second * 5)
}
