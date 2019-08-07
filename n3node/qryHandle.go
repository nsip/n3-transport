package n3node

import (
	"github.com/google/uuid"
	"github.com/nsip/n3-messages/messages/pb"
	"github.com/nsip/n3-transport/n3influx"
)

func mkTicket(dbClt *n3influx.DBClient, ctx, sub string, end, v int64) (ts []*pb.SPOTuple) {

	termID, endV := uuid.New().String(), INum2Str(I64(end))
	ts = append(ts, &pb.SPOTuple{Subject: sub, Predicate: termID, Object: endV, Version: v}) // *** return result ***

	// 	I := 0
	// AGAIN:
	// 	if _, ok := mTickets.Load(sub); ok {
	// 		time.Sleep(time.Millisecond * DELAY_CONTEST)
	// 		I++
	// 		if I >= 30000 {
	// 			errMsg := "THIS OBJECT IS OUTSTANDING"
	// 			ts = append(ts, &pb.SPOTuple{Subject: errMsg, Predicate: errMsg, Object: errMsg, Version: 999})
	// 			return
	// 		}
	// 		goto AGAIN
	// 	}

	// mTickets.Store(sub, &ticket{tktID: termID, idx: endV})
	// ctx = Str(ctx).MkSuffix("-meta").V()
	// GoFn("ticket", 1, false, ticketRmAsync, dbClt, &mTickets, ctx) // *** start infinite loop for deleting ticket ***

	return
}
