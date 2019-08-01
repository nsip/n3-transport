// influxpub.go

// publisher test for tuples to influx

package main

import (
	"log"
	"time"

	"github.com/nats-io/nuid"
	"github.com/nsip/n3-messages/messages"
	inf "github.com/nsip/n3-transport/n3influx"
)

// example tuple publisher for influx data-store
func main() {

	infPub, err := inf.NewPublisher()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 100; i++ {

		subject := "4BD6B062-66DD-474B-9E24-E3F85FB61FED" + nuid.Next()
		predicate := "TeachingGroup.TeachingGroupPeriodList.TeachingGroupPeriod[1].DayId" + nuid.Next()
		object := "F"
		context := "SIF"

		tuple, err := messages.NewTuple(subject, predicate, object)
		if err != nil {
			log.Fatal(err)
		}

		tuple.Version = 1 // arbitrary for testing

		err = infPub.StoreTuple(tuple, context)
		if err != nil {
			log.Fatal(err)
		}

	}

	// allow enough time for any last batches to be submitted
	log.Println("waiting...")
	time.Sleep(time.Millisecond * 2000)

}
