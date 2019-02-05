package n3node

import (
	"time"

	"../n3influx"
	u "github.com/cdutwhu/go-util"
	"github.com/nsip/n3-messages/messages/pb"
	"golang.org/x/sync/syncmap"
)

func getValueVerRange(dbClient *n3influx.DBClient, objID string, ctx string) (start, end, ver int64) {
	tuple := &pb.SPOTuple{Subject: objID, Predicate: "V"}
	o, v := dbClient.GetObjVer(tuple, u.Str(ctx).MkSuffix("-meta"))
	if v != -1 {
		ss := sSpl(o, "-")
		start, end, ver = u.Str(ss[0]).ToInt64(), u.Str(ss[1]).ToInt64(), v
	}
	return
}

func logMeta(dbClient *n3influx.DBClient, ctx, id string, start, end, verMeta int64) {
	// *** Save prevID's low-high version map into meta db as <"id" - "" - "low-high"> ***
	dbClient.StoreTuple(
		&pb.SPOTuple{
			Subject:   id,
			Predicate: "V",
			Object:    fSf("%d-%d", start, end),
			Version:   verMeta,
		},
		u.Str(ctx).MkSuffix("-meta")) // *** Meta Context ***
}

func ticketRmAsync(dbClient *n3influx.DBClient, tkts *syncmap.Map, ctx string) {
	ctx = u.Str(ctx).RmSuffix("-meta")
	for {
		tkts.Range(func(k, v interface{}) bool {
			if o, _ := dbClient.GetObjVer(&pb.SPOTuple{Subject: v.(*ticket).tktID, Predicate: TERMMARK}, ctx); o == k {
				fPln(k, "pub done!")
				tkts.Delete(k)
			}
			return true // *** continue range ***
		})
		time.Sleep(time.Millisecond * DELAY_CHKTERM)
	}
}

func assignVer(dbClient *n3influx.DBClient, tuple *pb.SPOTuple, ctx string) bool {
	s, p, o, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version

	// *** for value tuple ***
	if u.Str(s).IsUUID() {

		// *** New ID (NOT Terminator) is coming ***
		if s != prevID && p != TERMMARK {

			// *** Put incoming version into its own queue ***
			mapIDVQueue[s] = append(mapIDVQueue[s], v)
			l := len(mapIDVQueue[s])
			startVer = mapIDVQueue[s][l-1]
		}

		// *** Terminator is coming, ready to log a meta ***
		if p == TERMMARK {
			mapVerToMeta.Store(prevID, &metaData{ID: prevID, StartVer: startVer, EndVer: prevVer, Ver: verMeta})
			verMeta++
		}

		prevID, prevPred, prevVer = s, p, v
	}

	// *** check struct tuple ***
	if p == "::" {
		if objDB, verDB := dbClient.GetObjVer(tuple, ctx); verDB > 0 {
			if u.Str(objDB).FieldsSeqContain(o, " + ") {
				tuple.Version = 0
				return false
			}
		}
	}

	return true
}

// inDB : is before db storing
func inDB(dbClient *n3influx.DBClient, tuple *pb.SPOTuple, ctx string) bool {
	s, p, o, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version

	if u.Str(s).IsUUID() && p != TERMMARK {

		// *** when n3node is restarting, fetch check version from meta data ***
		if _, ok := mapVerInDBChk[s]; !ok {
			_, mapVerInDBChk[s], _ = getValueVerRange(dbClient, s, ctx)
		}
		if v <= mapVerInDBChk[s] {
			fPln(v, mapVerInDBChk[s])
			return true
		}

		// *** Save prevID's low-high version map into meta db as <"id" - "V" - "low-high"> ***
		if value, ok := mapVerToMeta.Load(s); ok {
			md := value.(*metaData)
			logMeta(dbClient, ctx, s, md.StartVer, md.EndVer, md.Ver)
			mapVerToMeta.Delete(s)
		}
	}

	if p == "::" || u.Str(p).IsUUID() || p == TERMMARK {
		if objDB, verDB := dbClient.GetObjVer(tuple, ctx); verDB > 0 {
			return o == objDB
		}
	}

	return false
}
