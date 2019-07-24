package n3influx

import (
	"testing"

	"github.com/nsip/n3-messages/messages/pb"
)

func TestDbTblExists(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	fPln(dbClient.DbTblExists("measurements", "ctxid"))
}

func TestLatestVer(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	fPln(dbClient.LatestVer("ctxid"))
}

func TestRootByID(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	fPln(dbClient.RootByID("demo", "00e6291b-884f-4fa2-92a1-bd477634be97", " ~ "))
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

func TestIDListByRoot(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)

	ids := dbClient.IDListByRoot("demo", "xapi")
	for i, id := range ids {
		fPln(i, ":", id)
	}
	fPln(" ids ------------------------------------- ")
}

func TestIDListByPathValue(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)

	tuple11 := &pb.SPOTuple{Predicate: MARKTerm, Object: "1822AF7A-F9CB-4F0D-96EA-9280DD0B6AB2"}
	ids11 := dbClient.IDListByPathValue("demo", tuple11, true)
	for i, id := range ids11 {
		fPln(i, ":", id)
	}

	_, _, ids22, _, _ := dbClient.OsBySP("demo", "644153cf-02c2-4670-810b-534ac148c011", MARKTerm, false, false, 0, 0)
	for i, id := range ids22 {
		fPln(i, ":", id)
	}

	tuple := &pb.SPOTuple{Predicate: "NIAS3 ~ actor ~ name", Object: "Lillian Simon"}
	ids1 := dbClient.IDListByPathValue("demo", tuple, false)
	for i, id := range ids1 {
		fPln(i, ":", id)
	}
	fPln(" ids1 ------------------------------------- ")

	tuple = &pb.SPOTuple{Predicate: "NIAS3 ~ object ~ id", Object: "http://example.com/assignments/Geography-7-1-B:4"}
	ids2 := dbClient.IDListByPathValue("demo", tuple, true)
	for i, id := range ids2 {
		fPln(i, ":", id)
	}
	fPln(" ids2 ------------------------------------- ")

	return

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

func TestSingleListOfSPO(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	for i, r := range dbClient.SingleListOfSPO("ctxid", "s") { // ****** should list all s-p pairs
		fPln(i, r)
		fPln(dbClient.LastPOByS("ctxid", r))
	}
}

func TestLastPOByS(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	fPln(dbClient.POsByS("ctxid", "xapi", "", MARKDelID, 0, 0))
}

func TestLastObySP(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	fPln(dbClient.LastOBySP("ctxid", "xapi", "demo"))
}

func TestPairListOfSPO(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	rst1, rst2, rstLeft := dbClient.PairListOfSPO("ctxid", "O")
	for i := 0; i < len(rst1); i++ {
		if rstLeft[i] != MARKDelID {
			fPln(rst1[i], rst2[i], rstLeft[i])
		}
	}
}
