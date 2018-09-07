// tmpub.go test publisher for tendermint

package main

import (
	"log"
	"time"

	"github.com/nats-io/nuid"
	"github.com/nsip/n3-transport/messages"
	tm "github.com/nsip/n3-transport/n3tendermint"
)

// example tuple publisher for tendermint transport
func main() {

	tmPub, err := tm.NewPublisher()
	if err != nil {
		log.Println(err)
	}

	for i := 0; i < 100; i++ {

		subject := "4BD6B062-66DD-474B-9E24-E3F85FB61FED" + nuid.Next()
		predicate := "TeachingGroup.TeachingGroupPeriodList.TeachingGroupPeriod[1].DayId" + nuid.Next()
		object := "F"
		context := "SIF"

		tuple, err := messages.NewTuple(subject, predicate, object, context)
		if err != nil {
			log.Fatal(err)
		}
		tuple.Version = 1 // arbitrary for testing

		msg, err := messages.NewMessage(tuple)
		if err != nil {
			log.Fatal(err)
		}

		err = tmPub.SubmitTx(msg)
		if err != nil {
			log.Println("tx error: ", err)
		} else {
			// log.Println("tx successfully submitted ", i)
		}

	}

	// allow a small amount of catch-up time before quitting
	log.Println("waiting...")
	time.Sleep(time.Millisecond * 2000)

}
