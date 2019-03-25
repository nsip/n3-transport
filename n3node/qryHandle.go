package n3node

import (
	"time"

	"../n3influx"
	u "github.com/cdutwhu/go-util"
	"github.com/google/uuid"
	"github.com/nsip/n3-messages/messages/pb"
)

// func queryHandle(dbClt *n3influx.DBClient, tuple *pb.SPOTuple, ctx, pathDel, childDel string, start, end int64) (ts []*pb.SPOTuple) {

// 	tempCtx := fSf("temp_%d", time.Now().UnixNano())
// 	dbClt.BatTrans(tuple, ctx, tempCtx, false, true, start, end)

// 	tupleS := &pb.SPOTuple{Subject: tuple.Predicate, Predicate: "::"}
// 	dbClt.BatTrans(tupleS, ctx, tempCtx, true, false, 0, 0)

// 	tupleA := &pb.SPOTuple{Subject: tuple.Predicate, Predicate: tuple.Subject}
// 	dbClt.BatTrans(tupleA, ctx, tempCtx, true, false, 0, 0)

// 	/******************************************/
// 	arrInfo := make(map[string]int) /* key: predicate, value: array count */
// 	// ctx, revArr := n3msg.CtxName, true /* Search from original measurement, reverse array order */
// 	// dbClt.QueryTuple(tuple, 0, revArr, ctx, &ts, &arrInfo, start, end) /* Search from original measurement */
// 	ctx, revArr := tempCtx, true                                                    /* Search from temp measurement, reverse array order */
// 	dbClt.QueryTuple(tuple, 0, revArr, ctx, pathDel, childDel, &ts, &arrInfo, 0, 0) /* Search from temp measurement */
// 	dbClt.AdjustOptionalTuples(&ts, &arrInfo)                                       /* we need re-order some tuples */

// 	/******************************************/
// 	dbClt.DropCtx(tempCtx)

// 	return
// }

func requestTicket(dbClt *n3influx.DBClient, ctx, sub string, end, v int64) (ts []*pb.SPOTuple) {

AGAIN:
	if _, ok := mapTickets.Load(sub); ok {
		time.Sleep(time.Millisecond * DELAY_CONTEST)
		goto AGAIN
	}

	termID, endV := uuid.New().String(), u.I64(end).ToStr()
	ts = append(ts, &pb.SPOTuple{Subject: sub, Predicate: termID, Object: endV, Version: v}) // *** return result ***

	mapTickets.Store(sub, &ticket{tktID: termID, idx: endV})
	u.GoFn("ticket", 1, false, ticketRmAsync, dbClt, &mapTickets, ctx)

	return
}
