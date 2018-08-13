// pub.go

package main

import (
	"log"
	"time"

	"github.com/nats-io/nuid"
	tm "github.com/nsip/n3-transport/n3tendermint"
)

// example tuple publisher
func main() {

	n3c, err := tm.NewPublisher()
	if err != nil {
		log.Println(err)
	}
	// _ = n3c

	// n3ic, err := inf.NewPublisher("SIF2")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	memMap := make(map[string]int64)

	for i := 0; i < 10000; i++ {

		subject := "4BD6B062-66DD-474B-9E24-E3F85FB61FED" + nuid.Next()
		predicate := "TeachingGroup.TeachingGroupPeriodList.TeachingGroupPeriod[1].DayId"
		object := "F"
		context := "SIF"

		key := subject + predicate
		memMap[key]++
		version := memMap[key]

		// tuple := &pb.SPOTuple{
		// 	Subject:   subject,
		// 	Predicate: predicate,
		// 	Object:    object,
		// 	Version:   version,
		// }

		// err := n3ic.StoreTuple(tuple)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		msg, err := n3c.NewMessage(subject, predicate, object, context, version)
		if err != nil {
			panic(err)
		}

		// log.Println("version: ", version)

		err = n3c.SubmitTx(msg)
		if err != nil {
			log.Println("tx error: ", err)
		} else {
			// log.Println("tx successfully submitted ", i)
		}

		// if (i % 5000) == 0 {
		// 	time.Sleep(time.Millisecond * 500)
		// }

	}

	log.Println("waiting...")
	time.Sleep(time.Millisecond * 2000)

}
