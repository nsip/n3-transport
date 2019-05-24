package n3node

import (
	"time"

	"../n3influx"
	"github.com/nsip/n3-messages/messages/pb"
	"golang.org/x/sync/syncmap"
)

func getVerRange(dbClient *n3influx.DBClient, objID string, ctx, metaType string) (start, end, ver int64) {
	if exist, alive := dbClient.Status(objID, ctx); exist && alive {
		o, v := dbClient.LastObjVer(&pb.SPOTuple{Subject: objID, Predicate: metaType}, Str(ctx).MkSuffix("-meta").V())
		if v != -1 {
			ss := sSpl(o, "-")
			start, end, ver = Str(ss[0]).ToInt64(), Str(ss[1]).ToInt64(), v
		}
	} else if !exist {
		start, end, ver = 0, 0, 0
	} else if exist && !alive {
		start, end, ver = -1, -1, -1
	}
	return
}

func mkMetaTuple(dbClient *n3influx.DBClient, ctx, id string, start, end int64, metaType string) (*pb.SPOTuple, string) {
	ctxMeta := Str(ctx).MkSuffix("-meta").V()
	_, v := dbClient.LastObjVer(&pb.SPOTuple{Subject: id, Predicate: metaType}, ctxMeta)
	return &pb.SPOTuple{
			Subject:   id,
			Predicate: metaType,
			Object:    fSf("%d-%d", start, end),
			Version:   IF(v == -1, int64(1), v+1).(int64),
		},
		ctxMeta // *** Meta Context ***
}

// ticketRmAsync : args : *n3influx.DBClient, *syncmap.Map, string
func ticketRmAsync(done <-chan int, id int, args ...interface{}) {
	dbClient, tkts, ctx := args[0].(*n3influx.DBClient), args[1].(*syncmap.Map), args[2].(string)
	ctx = Str(ctx).RmSuffix("-meta").V()
	i := 0
	for {
		bInRange := false
		i++
		tkts.Range(func(k, v interface{}) bool {
			bInRange = true
			if o, _ := dbClient.LastObjVer(&pb.SPOTuple{Subject: v.(*ticket).tktID, Predicate: TERMMARK}, ctx); o == k {
				tkts.Delete(k)
			} else {
				fPf("there is an outstanding@%6d : %s - %s. waiting...\n", i, k, v.(*ticket).tktID)
				time.Sleep(time.Millisecond * 1000)
			}
			return true //                                                          *** continue range ***
		})
		time.Sleep(time.Millisecond * DELAY_CHKTERM)
		if !bInRange {
			// fPln("pub done !")
		}
	}
	<-done
}

// assignVer : continue to save, additional tuple, additional context
func assignVer(dbClient *n3influx.DBClient, tuple *pb.SPOTuple, ctx string) (metaTuple *pb.SPOTuple, metaCtx string) {

	s, p, o, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version
	fPln("assignVer:", s, p, o, v)

	if Str(s).IsUUID() && !Str(o).HP("::") && !Str(o).HP("[]") { //            *** value tuple ***

		if s != prevIDv && p != TERMMARK { //         *** New ID (NOT Terminator) is coming ***
			mIDvQueue[s] = append(mIDvQueue[s], v) // *** Put incoming version into its own queue ***
			l := len(mIDvQueue[s])
			startVer = mIDvQueue[s][l-1]
		}
		if p == TERMMARK { //                         *** Terminator, ready to create a meta tuple ***
			metaTuple, metaCtx = mkMetaTuple(dbClient, ctx, prevIDv, startVer, prevVerV, "V")
		}
		prevIDv, prevVerV = s, v

	} else if Str(p).HP("::") || (Str(s).IsUUID() && Str(o).HP("::")) { //     *** struct tuple ***

		if p != prevIDs && p != TERMMARK {
			mIDsQueue[p] = append(mIDsQueue[p], v)
			l := len(mIDsQueue[p])
			startVer = mIDsQueue[p][l-1]
		}
		if p == TERMMARK {
			metaTuple, metaCtx = mkMetaTuple(dbClient, ctx, prevIDs, startVer, prevVerS, "S")
		}
		prevIDs, prevVerS = p, v

	} else if Str(p).HP("[]") || (Str(s).IsUUID() && Str(o).HP("[]")) { //     *** array tuple ***

		if p != prevIDa && p != TERMMARK {
			mIDaQueue[p] = append(mIDaQueue[p], v)
			l := len(mIDaQueue[p])
			startVer = mIDaQueue[p][l-1]
		}
		if p == TERMMARK {
			metaTuple, metaCtx = mkMetaTuple(dbClient, ctx, prevIDa, startVer, prevVerA, "A")
		}
		prevIDa, prevVerA = p, v
	}

	return
}

// inDB : before db storing, if Object is "deleted", it's not inDB
func inDB(dbClient *n3influx.DBClient, tuple *pb.SPOTuple, ctx string) bool {
	s, p, _, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version

	if !Str(ctx).HS("-meta") { //                                                           *** DATA TABLE ***

		if Str(s).IsUUID() && p != TERMMARK { //                                            *** VALUES ***
			if _, end, _ := getVerRange(dbClient, s, ctx, "V"); v <= end {
				return true
			}
		} else if Str(p).HP("::") { //                                                      *** STRUCT ***
			if _, end, _ := getVerRange(dbClient, p, ctx, "S"); v <= end {
				return true
			}
		} else if Str(p).HP("[]") { //                                                      *** ARRAY ***
			if _, end, _ := getVerRange(dbClient, p, ctx, "A"); v <= end {
				return true
			}
		} else { //                                                                         *** TERMMARK ***
			return dbClient.TupleExists(tuple, ctx, "Subject")
		}

	} else { //                                                                             *** META TABLE ***

		return dbClient.TupleExists(tuple, ctx)
	}

	return false
}
