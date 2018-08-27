// influxpub.go

// publisher test for tuples to influx

package main

import (
	"log"
	"time"

	"github.com/nats-io/nuid"
	inf "github.com/nsip/n3-transport/n3influx"
	"github.com/nsip/n3-transport/pb"
)

// example tuple publisher for influx data-store
func main() {

	infPub, err := inf.NewPublisher()
	if err != nil {
		log.Fatal(err)
	}

	memMap := make(map[string]int64)

	for i := 0; i < 1000; i++ {

		subject := "4BD6B062-66DD-474B-9E24-E3F85FB61FED" + nuid.Next()
		predicate := "TeachingGroup.TeachingGroupPeriodList.TeachingGroupPeriod[1].DayId" + nuid.Next()
		object := "F"
		context := "SIF"

		key := subject + predicate
		memMap[key]++
		version := memMap[key]

		tuple := &pb.SPOTuple{
			Subject:   subject,
			Predicate: predicate,
			Object:    object,
			Version:   version,
			Context:   context,
		}

		err := infPub.StoreTuple(tuple)
		if err != nil {
			log.Fatal(err)
		}

	}

	// allow enough time for any last batches to be submitted
	log.Println("waiting...")
	time.Sleep(time.Millisecond * 2000)

}
