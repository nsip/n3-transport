package n3influx

import (
	"testing"
	"time"

	"github.com/nsip/n3-messages/messages/pb"
)

func TestStoreTuple(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)

	tuple := &pb.SPOTuple{
		Subject:   "demo",
		Predicate: "r/w",
		Object:    "StaffPersonal ~ PersonInfo",
	}
	pe(dbClient.StoreTuple(tuple, "blacklist"))

	time.Sleep(1 * time.Second)
}
