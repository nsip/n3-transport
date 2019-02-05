package n3influx

import (
	"encoding/json"
	"strings"

	u "github.com/cdutwhu/go-util"
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/nsip/n3-messages/messages/pb"
)

// SubExist :
func (n3ic *DBClient) SubExist(tuple *pb.SPOTuple, ctx string, vLow, vHigh int64) bool {
	// pln("checking subject ...")

	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' AND tombstone='false' `+vChkL+vChkH, tuple.Subject)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())
	return len(resp.Results[0].Series) != 0
}

// BatTransEx :
func (n3ic *DBClient) BatTransEx(tuple *pb.SPOTuple, ctx, ctxNew string, extSub, extPred bool, vLow, vHigh int64,
	exclude func(s, p, o string, v int64) bool) (n int64) {

	if ss, ps, os, vs, ok := n3ic.GetObjs(tuple, ctx, extSub, extPred, vLow, vHigh); ok {
		for i := range ss {
			if exclude(ss[i], ps[i], os[i], vs[i]) {
				continue
			}
			temp := &pb.SPOTuple{Subject: ss[i], Predicate: ps[i], Object: os[i], Version: vs[i]}
			PE(n3ic.StoreTuple(temp, ctxNew))
			n++
		}
	}
	return n
}

// BatTrans :
func (n3ic *DBClient) BatTrans(tuple *pb.SPOTuple, ctx, ctxNew string, extSub, extPred bool, vLow, vHigh int64) int64 {

	s, p := tuple.Subject, tuple.Predicate
	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect, qWhere := fSf(`SELECT version, subject, predicate, object, tombstone INTO "%s" FROM "%s" `, ctxNew, ctx), ""
	if extSub && !extPred {
		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate='%s' AND tombstone='false' `+vChkL+vChkH, s, p)
	} else if extSub && extPred {
		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate=~/^%s/ AND tombstone='false' `+vChkL+vChkH, s, p)
	} else if !extSub && extPred {
		qWhere = fSf(`WHERE subject='%s' AND predicate=~/^%s/ AND tombstone='false' `+vChkL+vChkH, s, p)
	} else if !extSub && !extPred {
		qWhere = fSf(`WHERE subject='%s' AND predicate='%s' AND tombstone='false' `+vChkL+vChkH, s, p)
	}
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)

	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())

	return Must(resp.Results[0].Series[0].Values[0][1].(json.Number).Int64()).(int64)
}

// GetObjs (for XAPI query) : (return objects, versions, IsFound)
func (n3ic *DBClient) GetObjs(tuple *pb.SPOTuple, ctx string, extSub, extPred bool, vLow, vHigh int64) (subs, preds, objs []string, vers []int64, found bool) {

	s, p := tuple.Subject, tuple.Predicate
	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect, qWhere := fSf(`SELECT subject, predicate, object, version FROM "%s" `, ctx), ""
	if extSub && !extPred {
		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate='%s' AND tombstone='false' `+vChkL+vChkH, s, p)
	} else if extSub && extPred {
		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate=~/^%s/ AND tombstone='false' `+vChkL+vChkH, s, p)
	} else if !extSub && extPred {
		qWhere = fSf(`WHERE subject='%s' AND predicate=~/^%s/ AND tombstone='false' `+vChkL+vChkH, s, p)
	} else if !extSub && !extPred {
		qWhere = fSf(`WHERE subject='%s' AND predicate='%s' AND tombstone='false' `+vChkL+vChkH, s, p)
	}
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)

	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())

	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {
			sub, pred, obj := l[1].(string), l[2].(string), l[3].(string)
			subs, preds, objs = append(subs, sub), append(preds, pred), append(objs, obj)
			v := Must(l[4].(json.Number).Int64()).(int64)
			ver := v
			vers = append(vers, ver)
			// fPln(pred, obj, ver)
		}
		found = true
	}
	return
}

// QueryTuples (for XAPI query) :
func (n3ic *DBClient) QueryTuples(tuple *pb.SPOTuple, ctx string, ts *[]*pb.SPOTuple, vLow, vHigh int64) {
	if !n3ic.SubExist(tuple, ctx, vLow, vHigh) {
		fPln("subject does not exist !")
		return
	}
	if subs, preds, objs, vers, ok := n3ic.GetObjs(tuple, ctx, false, true, vLow, vHigh); ok {
		for i := range subs {
			*ts = append(*ts, &pb.SPOTuple{
				Subject:   subs[i],
				Predicate: preds[i],
				Object:    objs[i],
				Version:   vers[i],
			})
		}
	}
}

/**********************************************************************************************/

// GetObjVer :
func (n3ic *DBClient) GetObjVer(tuple *pb.SPOTuple, ctx string) (string, int64) {
	if obj, ver, found := n3ic.GetObj(tuple, 0, ctx, 0, 0); found {
		return obj, ver
	}
	return "", -1
}

// GetObj : (return object, version, IsFound)
func (n3ic *DBClient) GetObj(tuple *pb.SPOTuple, offset int, ctx string, vLow, vHigh int64) (obj string, ver int64, found bool) {

	ver = -1
	s, p := tuple.Subject, tuple.Predicate
	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' AND tombstone='false' `+vChkL+vChkH, s, p)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1 OFFSET %d`, orderByTm, offset)

	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		// pln(resp.Results[0].Series[0].Values)
		// pln(resp.Results[0].Series[0].Values[0])
		// pln(resp.Results[0].Series[0].Values[0][1]) /* [0] is time, [1] is object, as SELECT ... */

		obj = resp.Results[0].Series[0].Values[0][1].(string)
		v := Must(resp.Results[0].Series[0].Values[0][2].(json.Number).Int64()).(int64)
		ver, found = v, true
		return
	}
	return
}

// GetSubStruct :
func (n3ic *DBClient) getSubStruct(tuple *pb.SPOTuple, ctx string, vLow, vHigh int64) (string, bool) {
	// pln("looking for sub struct ...")

	s, p := tuple.Predicate, "::"
	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' AND version!=0 AND tombstone='false' `+vChkL+vChkH, s, p)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)

	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		return resp.Results[0].Series[0].Values[0][1].(string), true
	}
	return "", false
}

// GetArrInfo :
func (n3ic *DBClient) getArrInfo(tuple *pb.SPOTuple, ctx string, vLow, vHigh int64) (int, bool) {
	// pln("looking for array info ...")

	s, p := tuple.Predicate, tuple.Subject
	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' AND version!=0 AND tombstone='false' `+vChkL+vChkH, s, p)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)

	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		return u.Str(resp.Results[0].Series[0].Values[0][1].(string)).ToInt(), true
	}
	return 0, false
}

// QueryTuple :
func (n3ic *DBClient) QueryTuple(tuple *pb.SPOTuple, offset int, revArr bool, ctx string,
	ts *[]*pb.SPOTuple, arrInfo *map[string]int, vLow, vHigh int64) {

	if !n3ic.SubExist(tuple, ctx, vLow, vHigh) {
		fPln("subject does not exist !")
		return
	}

	if obj, ver, ok := n3ic.GetObj(tuple, offset, ctx, vLow, vHigh); ok {
		// fPf("got object .................. %d \n", ver)
		tuple.Object = obj
		*ts = append(*ts, &pb.SPOTuple{
			Subject:   tuple.Subject,
			Predicate: tuple.Predicate,
			Object:    obj,
			Version:   ver,
		})
		return
	}

	if stru, ok := n3ic.getSubStruct(tuple, ctx, 0, 0); ok {
		// fPf("got sub-struct, more work ------> %s\n", stru)

		/* Array Element */
		if i := sI(stru, "[]"); i == 0 {
			nArr, _ := n3ic.getArrInfo(tuple, ctx, 0, 0)
			// pln(nArr)
			tuple.Predicate += fSf(".%s", stru[i+2:])
			// pln(tuple.Predicate)
			(*arrInfo)[tuple.Predicate] = nArr /* keep the array */

			for k := 0; k < nArr; k++ {
				n3ic.QueryTuple(tuple, u.TerOp(revArr, nArr-k-1, k).(int), revArr, ctx, ts, arrInfo, vLow, vHigh)
			}
			tuple.Predicate = u.Str(tuple.Predicate).RmTailFromLast(".")
			return
		}

		/* None Array Element */
		subs := strings.Split(stru, " + ")
		for _, s := range subs {
			tuple.Predicate += fSf(".%s", s)
			// pln(tuple.Predicate)
			n3ic.QueryTuple(tuple, offset, revArr, ctx, ts, arrInfo, vLow, vHigh)
			tuple.Predicate = u.Str(tuple.Predicate).RmTailFromLast(".")
		}
		return

		/* Only for debugging */
		// tuple.Object = stru
		// *ts = append(*ts, tuple)
		// return
	}
}

type arrOptTupleInfo struct {
	nArray   int     /* 'this' tuple's array's count */
	nTuple   int     /* 'this' tuple count in real */
	indices  []int   /* optional tuple's index in ts */
	versions []int64 /* optional tuple's version */
}

func getVersBySPO(ts *[]*pb.SPOTuple, spo ...string) (vers []int64, indices []int) {
	s, p, o := spo[0], spo[1], spo[2]
	for i, t := range *ts {
		if (s != "" && p != "" && o != "") && (t.Subject == s && t.Predicate == p && t.Object == o) {
			goto APPEND
		} else if (s != "" && p != "" && o == "") && (t.Subject == s && t.Predicate == p) {
			goto APPEND
		} else if (s != "" && p == "" && o != "") && (t.Subject == s && t.Object == o) {
			goto APPEND
		} else if (s == "" && p != "" && o != "") && (t.Predicate == p && t.Object == o) {
			goto APPEND
		} else if (s != "" && p == "" && o == "") && (t.Subject == s) {
			goto APPEND
		} else if (s == "" && p != "" && o == "") && (t.Predicate == p) {
			goto APPEND
		} else if (s == "" && p == "" && o != "") && (t.Object == o) {
			goto APPEND
		} else {
			continue
		}
	APPEND:
		vers = append(vers, t.Version)
		indices = append(indices, i)
	}
	return
}

// AdjustOptionalTuples : according to tuple version, adjust some tuples' order in sub array tuple
func (n3ic *DBClient) AdjustOptionalTuples(ts *[]*pb.SPOTuple, arrInfo *map[string]int) {
	mapTuplesInArr := make(map[string]*arrOptTupleInfo)
	for ituple, tuple := range *ts {
		p, v := tuple.Predicate, tuple.Version
		if ok, nArr := u.Str(p).CoverAnyKeyInMapSI(*arrInfo); ok {
			// fPf("%d : %s : %d # %d\n", ituple, p, nArr, v)
			if _, ok := mapTuplesInArr[p]; !ok {
				mapTuplesInArr[p] = &arrOptTupleInfo{}
			}
			mapTuplesInArr[p].nArray = nArr
			mapTuplesInArr[p].nTuple++
			mapTuplesInArr[p].indices = append(mapTuplesInArr[p].indices, ituple)
			mapTuplesInArr[p].versions = append(mapTuplesInArr[p].versions, v)
		}
	}

	mapOptTuple := make(map[string]*arrOptTupleInfo)
	for k, v := range mapTuplesInArr {
		if v.nArray != v.nTuple {
			mapOptTuple[k] = v
		}
	}

	optTuples, aboveTuples := []*pb.SPOTuple{}, []*pb.SPOTuple{}
	for _, v := range mapOptTuple {
		for i, tIdx := range v.indices {
			this, above := (*ts)[tIdx], (*ts)[tIdx-1]
			// fPf("%d --- %v\n", tIdx, this)
			if vers, indicesTS := getVersBySPO(ts, "", above.Predicate, ""); len(vers) > 0 {
				tVer := v.versions[i]
				_, idx := u.I64(tVer).Nearest(vers...)
				posShouldAfter := indicesTS[idx]
				shouldAfter := (*ts)[posShouldAfter]
				// fPf("Should be after : %d --- %v\n", posShouldBeAfter, shouldAfter)
				optTuples, aboveTuples = append(optTuples, this), append(aboveTuples, shouldAfter)
			}
		}
	}

	/******************************************/

	tempts := make([]interface{}, len(*ts))
	for j, t := range *ts {
		tempts[j] = t
	}
	for i := 0; i < len(optTuples); i++ {
		tempts, _, _, _ = u.GArr(tempts).MoveItemAfter(
			func(move interface{}) bool { return move == optTuples[i] },
			func(after interface{}) bool { return after == aboveTuples[i] },
		)
	}
	for k, t := range tempts {
		(*ts)[k] = t.(*pb.SPOTuple)
	}
}
