package n3influx

import (
	"testing"
	"time"

	"github.com/nsip/n3-messages/messages/pb"
)

func TestStoreTuple(t *testing.T) {
	defer func() { PH(recover(), "./log.txt", true) }()
	dbClient, e := NewPublisher()
	PE(e)

	tuple := &pb.SPOTuple{
		Subject:   "D3E34F41-9D75-101A-8C3D-00AA001A1652",
		Predicate: "StaffPersonal.PersonInfo.AddressList.Address.-Type",
		Object:    "0123A111",
	}
	e = dbClient.StoreTuple(tuple, "temp")
	PE(e)

	time.Sleep(2 * time.Second)
}
