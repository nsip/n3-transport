package n3influx

import (
	"testing"
	"time"

	"github.com/nsip/n3-messages/messages/pb"
)

func TestGetObjs(t *testing.T) {
	defer func() { PH(recover(), "./log.txt", true) }()
	dbClient, e := NewPublisher()
	PE(e)

	tuple := &pb.SPOTuple{Subject: "D3E34F41-9D75-101A-8C3D-00AA001A1652", Predicate: "StaffPersonal"}
	// tuple := &pb.SPOTuple{Subject: "StaffPersonal", Predicate: "::"}
	if ss, ps, os, vs, ok := dbClient.GetObjs(tuple, "abc-sif", true, false, 0, 0); ok {
		for i := range ss {
			fPln(ss[i], ps[i], os[i], vs[i])
			/*************************************************/
			tuple1 := &pb.SPOTuple{
				Subject:   ss[i],
				Predicate: ps[i],
				Object:    os[i],
				Version:   vs[i],
			}
			PE(dbClient.StoreTuple(tuple1, "temp"))
		}
	}
	time.Sleep(20 * time.Millisecond)
}

func TestBatTrans(t *testing.T) {
	defer func() { PH(recover(), "./log.txt", true) }()
	dbClient, e := NewPublisher()
	PE(e)

	tuple := &pb.SPOTuple{Subject: "D3E34F41-9D75-101A-8C3D-00AA001A1652", Predicate: "StaffPersonal"}
	fPln(dbClient.BatTrans(tuple, "abc-sif", "temp1", false, true, 0, 0))
}

func TestBatTransEx(t *testing.T) {
	defer func() { PH(recover(), "./log.txt", true) }()
	dbClient, e := NewPublisher()
	PE(e)

	tuple := &pb.SPOTuple{Subject: "StaffPersonal", Predicate: "::"}
	fPln(dbClient.BatTransEx(tuple, "abc-sif", "temp2", true, false, 0, 0, func(s, p, o string, v int64) bool {
		return false
	}))
}
