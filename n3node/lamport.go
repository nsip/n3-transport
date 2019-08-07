package n3node

import (
	"github.com/nsip/n3-messages/messages/pb"
	"github.com/nsip/n3-transport/n3influx"
	// "golang.org/x/sync/syncmap"
)

func getVerRange(dbClient *n3influx.DBClient, ctx, objID, metaType string) (start, end, ver int64) {
	if exist, alive := dbClient.Status(ctx, objID); exist && alive {
		ctxMeta := S(ctx).MkSuffix("-meta").V()
		o, v := dbClient.LastOV(ctxMeta, &pb.SPOTuple{Subject: objID, Predicate: metaType})
		if v != -1 {
			ss := sSpl(o, "-")
			start, end, ver = S(ss[0]).ToInt64(), S(ss[1]).ToInt64(), v
		}
	} else if !exist {
		start, end, ver = 0, 0, 0
	} else if exist && !alive {
		start, end, ver = -1, -1, -1
	}
	return
}

func mkMetaTuple(dbClient *n3influx.DBClient, ctx, id string, start, end int64, metaType string) (*pb.SPOTuple, string) {
	ctxMeta := S(ctx).MkSuffix("-meta").V()
	_, v := dbClient.LastOV(ctxMeta, &pb.SPOTuple{Subject: id, Predicate: metaType})
	return &pb.SPOTuple{
			Subject:   id,
			Predicate: metaType,
			Object:    fSf("%d-%d", start, end),
			Version:   IF(v == -1, int64(1), v+1).(int64),
		},
		ctxMeta // *** Meta Context ***
}

// ticketRmAsync : args : *n3influx.DBClient, *syncmap.Map, string
// func ticketRmAsync(done <-chan int, id int, args ...interface{}) {
// 	dbClient, tkts, ctx := args[0].(*n3influx.DBClient), args[1].(*syncmap.Map), args[2].(string)
// 	ctx = Str(ctx).RmSuffix("-meta").V()
// 	i := 0
// 	for {
// 		bInRange := false
// 		i++
// 		tkts.Range(func(k, v interface{}) bool {
// 			bInRange = true
// 			if o, _ := dbClient.LastObjVer(&pb.SPOTuple{Subject: v.(*ticket).tktID, Predicate: MARKTerm}, ctx); o == k {
// 				tkts.Delete(k)
// 			} else {
// 				fPf("there is an outstanding@%6d : %s - %s - %s. waiting...\n", i, k, o, v.(*ticket).tktID)
// 				// time.Sleep(time.Millisecond * 1000)
// 			}
// 			return true //                                                          *** continue range ***
// 		})
// 		time.Sleep(time.Millisecond * DELAY_CHKTERM)
// 		if !bInRange {
// 			// fPln("pub done !")
// 		}
// 	}
// 	<-done
// }

// assignVer : continue to save, additional tuple, additional context
func assignVer(dbClient *n3influx.DBClient, tuple *pb.SPOTuple, ctx string) (metaTuple *pb.SPOTuple, metaCtx string) {

	s, p, o, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version
	// fPln("assignVer:", s, p, o, v)

	if S(s).IsUUID() && !S(o).HP("::") && !S(o).HP("[]") { //        *** value tuple *** (exclude S & A terminator)

		if s != prevIDv && p != MARKTerm { //         *** New ID (NOT Terminator) is coming ***
			mIDvQueue[s] = append(mIDvQueue[s], v) // *** Put incoming version into its own queue ***
			l := len(mIDvQueue[s])
			startVer = mIDvQueue[s][l-1]
		}
		if p == MARKTerm { //                         *** Terminator, ready to create a meta tuple ***
			metaTuple, metaCtx = mkMetaTuple(dbClient, ctx, prevIDv, startVer, prevVerV, "V")
		}
		prevIDv, prevVerV = s, v

	} else if S(s).HP("::") || (S(s).IsUUID() && S(o).HP("::")) { // *** struct tuple *** (include S terminator)

		if s != prevIDs && p != MARKTerm {
			mIDsQueue[s] = append(mIDsQueue[s], v)
			l := len(mIDsQueue[s])
			startVer = mIDsQueue[s][l-1]
		}
		if p == MARKTerm {
			metaTuple, metaCtx = mkMetaTuple(dbClient, ctx, prevIDs, startVer, prevVerS, "S")
		}
		prevIDs, prevVerS = s, v

	} else if S(s).HP("[]") || (S(s).IsUUID() && S(o).HP("[]")) { // *** array tuple *** (include A terminator)

		if s != prevIDa && p != MARKTerm {
			mIDaQueue[s] = append(mIDaQueue[s], v)
			l := len(mIDaQueue[s])
			startVer = mIDaQueue[s][l-1]
		}
		if p == MARKTerm {
			metaTuple, metaCtx = mkMetaTuple(dbClient, ctx, prevIDa, startVer, prevVerA, "A")
		}
		prevIDa, prevVerA = s, v

	} else {

	}

	return
}

func inDB(dbClient *n3influx.DBClient, ctx string, tuple *pb.SPOTuple) bool {
	if dbClient.TupleExists(ctx, tuple) {
		return true
	}

	switch {
	case IArrEleIn(ctx, Ss{"ctxid", "privctrl"}):
		return false
	case S(ctx).HS("-meta"):
		return false
	default:
		return false
	}
}

// inDB : before db storing, if Object is "deleted", it's not inDB
// func inDB(dbClient *n3influx.DBClient, tuple *pb.SPOTuple, ctx string) bool {
// 	s, p, _, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version

// 	if !S(ctx).HS("-meta") { //                                                           *** DATA TABLE ***

// 		if S(s).IsUUID() && p != MARKTerm { //                                            *** VALUES ***
// 			if _, end, _ := getVerRange(dbClient, ctx, s, "V"); v <= end {
// 				return true
// 			}
// 		} else if S(p).HP("::") { //                                                      *** STRUCT ***
// 			if _, end, _ := getVerRange(dbClient, ctx, p, "S"); v <= end {
// 				return true
// 			}
// 		} else if S(p).HP("[]") { //                                                      *** ARRAY ***
// 			if _, end, _ := getVerRange(dbClient, ctx, p, "A"); v <= end {
// 				return true
// 			}
// 		} else { //                                                                       *** MARKTerm ***
// 			return dbClient.TupleExists(ctx, tuple, "Subject")
// 		}

// 	} else { //                                                                           *** META TABLE ***

// 		return dbClient.TupleExists(ctx, tuple)
// 	}

// 	return false
// }
