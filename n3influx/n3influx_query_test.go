package n3influx

import (
	"testing"

	"github.com/nsip/n3-messages/messages/pb"
)

func TestRootByID(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	fPln(dbClient.RootByID("00e6291b-884f-4fa2-92a1-bd477634be97", "demo", " ~ "))
}

func TestObjectCount(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	n := dbClient.ObjectCount("demo", "id")
	fPln("object count:", n)
}

// func TestGetObjs(t *testing.T) {
// 	defer func() { PH(recover(), "./log.txt") }()
// 	dbClient := must(NewDBClient()).(*DBClient)

// 	// tuple := &pb.SPOTuple{Subject: "D3E34F41-9D75-101A-8C3D-00AA001A1656", Predicate: "StaffPersonal"}
// 	tuple := &pb.SPOTuple{Subject: "StaffPersonal", Predicate: "::"}
// 	if ss, ps, os, vs, ok := dbClient.GetObjs(tuple, "demo", true, false, 0, 0); ok {
// 		for i := range ss {
// 			fPln(ss[i], ps[i], os[i], vs[i])
// 			/*************************************************/
// 			tuple1 := &pb.SPOTuple{
// 				Subject:   ss[i],
// 				Predicate: ps[i],
// 				Object:    os[i],
// 				Version:   vs[i],
// 			}
// 			PE(dbClient.StoreTuple(tuple1, "temp"))
// 		}
// 	}
// 	time.Sleep(20 * time.Millisecond)
// }

func TestIDListByPathValue(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)

	tuple := &pb.SPOTuple{Predicate: "NIAS3 ~ actor ~ name", Object: "Lillian Simon"}
	ids1 := dbClient.IDListByPathValue(tuple, "abc-xapi", false)
	// for i, id := range ids1 {
	// 	fPln(i, ":", id)
	// }
	// fPln(" ids1 ------------------------------------- ")

	tuple = &pb.SPOTuple{Predicate: "NIAS3 ~ object ~ id", Object: "http://example.com/assignments/Geography-7-1-B:4"}
	ids2 := dbClient.IDListByPathValue(tuple, "abc-xapi", false)
	// for i, id := range ids2 {
	// 	fPln(i, ":", id)
	// }
	// fPln(" ids2 ------------------------------------- ")

	ids := IArrIntersect(Ss(ids1), Ss(ids2))
	for i, id := range ids.([]string) {
		fPln(i, ":", id)
	}
}

// func TestBatTrans(t *testing.T) {
// 	defer func() { PH(recover(), "./log.txt") }()
// 	dbClient := Must(NewDBClient()).(*DBClient)
// 	tuple := &pb.SPOTuple{Subject: "D3E34F41-9D75-101A-8C3D-00AA001A1656", Predicate: "StaffPersonal"}
// 	fPln(dbClient.BatTrans(tuple, "abc-sif", "temp1", false, true, 0, 0))
// }

// func TestBatTransEx(t *testing.T) {
// 	defer func() { PH(recover(), "./log.txt") }()
// 	dbClient := Must(NewDBClient()).(*DBClient)
// 	tuple := &pb.SPOTuple{Subject: "StaffPersonal", Predicate: "::"}
// 	fPln(dbClient.BatTransEx(tuple, "abc-sif", "temp2", true, false, 0, 0, func(s, p, o string, v int64) bool {
// 		return false
// 	}))
// }

func TestAllSubs(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)

	fPln(dbClient.AllSubs("ctxid"))
}

func TestLastPredObjByS(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)

	fPln(dbClient.LastPredObjByS("ctxid", "xapi111"))
}