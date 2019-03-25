package n3influx

import (
	"encoding/json"

	u "github.com/cdutwhu/go-util"
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/nsip/n3-messages/messages/pb"
)

// RootByID :
func (n3ic *DBClient) RootByID(objID, ctx, del string) (root string) {
	if !sC(ctx, "-meta") && u.Str(objID).IsUUID() {
		if p, _, ok := n3ic.SubExist(&pb.SPOTuple{Subject: objID}, ctx, 0, 0); ok && sC(p, del) {
			root = sSpl(p, del)[0]
		}
	}
	return
}

// SubExist :
func (n3ic *DBClient) SubExist(tuple *pb.SPOTuple, ctx string, vLow, vHigh int64) (pred, obj string, exist bool) {
	// pln("checking subject ...")

	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect := fSf(`SELECT version, predicate, object FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' `+vChkL+vChkH, tuple.Subject)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())

	// pln(resp.Results[0].Series[0].Values[0][1]) /* [0] is time, [1] is as SELECT ... */
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		last := resp.Results[0].Series[0].Values[0]
		pred, obj, exist = last[2].(string), last[3].(string), true
	}

	return
}

// IDListByPathValue :
func (n3ic *DBClient) IDListByPathValue(tuple *pb.SPOTuple, ctx string) (ids []string) {

	path, value := tuple.Predicate, tuple.Object
	qSelect := fSf(`SELECT version, subject FROM "%s" `, ctx)
	qWhere := fSf(`WHERE predicate='%s' AND object='%s' `, path, value)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())

	if len(resp.Results[0].Series) > 0 {
		for _, v := range resp.Results[0].Series[0].Values {
			id := v[2].(string)
			if u.Str(id).IsUUID() {
				ids = append(ids, id)
			}
		}
	}

	return
}

// BatTransEx :
// func (n3ic *DBClient) BatTransEx(tuple *pb.SPOTuple, ctx, ctxNew string, extSub, extPred bool, vLow, vHigh int64,
// 	exclude func(s, p, o string, v int64) bool) (n int64) {

// 	if ss, ps, os, vs, ok := n3ic.GetObjs(tuple, ctx, extSub, extPred, vLow, vHigh); ok {
// 		for i := range ss {
// 			if exclude(ss[i], ps[i], os[i], vs[i]) {
// 				continue
// 			}
// 			temp := &pb.SPOTuple{Subject: ss[i], Predicate: ps[i], Object: os[i], Version: vs[i]}
// 			PE(n3ic.StoreTuple(temp, ctxNew))
// 			n++
// 		}
// 	}
// 	return n
// }

// BatTrans :
// func (n3ic *DBClient) BatTrans(tuple *pb.SPOTuple, ctx, ctxNew string, extSub, extPred bool, vLow, vHigh int64) int64 {

// 	s, p := tuple.Subject, tuple.Predicate
// 	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
// 	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

// 	qSelect, qWhere := fSf(`SELECT version, subject, predicate, object INTO "%s" FROM "%s" `, ctxNew, ctx), ""
// 	if extSub && !extPred {
// 		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate='%s' `+vChkL+vChkH, s, p)
// 	} else if extSub && extPred {
// 		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate=~/^%s/ `+vChkL+vChkH, s, p)
// 	} else if !extSub && extPred {
// 		qWhere = fSf(`WHERE subject='%s' AND predicate=~/^%s/ `+vChkL+vChkH, s, p)
// 	} else if !extSub && !extPred {
// 		qWhere = fSf(`WHERE subject='%s' AND predicate='%s' `+vChkL+vChkH, s, p)
// 	}
// 	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)

// 	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
// 	PE(e, resp.Error())

// 	return Must(resp.Results[0].Series[0].Values[0][1].(json.Number).Int64()).(int64)
// }

// GetObjs (for XAPI query) : (return objects, versions, IsFound)
func (n3ic *DBClient) GetObjs(tuple *pb.SPOTuple, ctx string, extSub, extPred bool, vLow, vHigh int64) (subs, preds, objs []string, vers []int64, found bool) {

	s, p := tuple.Subject, tuple.Predicate
	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect, qWhere := fSf(`SELECT subject, predicate, object, version FROM "%s" `, ctx), ""
	if extSub && !extPred {
		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate='%s' `+vChkL+vChkH, s, p)
	} else if extSub && extPred {
		qWhere = fSf(`WHERE subject=~/^%s/ AND predicate=~/^%s/ `+vChkL+vChkH, s, p)
	} else if !extSub && extPred {
		qWhere = fSf(`WHERE subject='%s' AND predicate=~/^%s/ `+vChkL+vChkH, s, p)
	} else if !extSub && !extPred {
		qWhere = fSf(`WHERE subject='%s' AND predicate='%s' `+vChkL+vChkH, s, p)
	}
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)

	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	PE(e, resp.Error())

	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {

			// if l[1] == nil || l[2] == nil || l[3] == nil {
			// 	fPln("***", l)
			// }

			sub, pred, obj := l[1].(string), l[2].(string), l[3].(string)
			subs, preds, objs = append(subs, sub), append(preds, pred), append(objs, obj)
			vers = append(vers, Must(l[4].(json.Number).Int64()).(int64))
			// fPln(pred, obj, ver)
		}
		found = true
	}
	return
}

// QueryTuples :
func (n3ic *DBClient) QueryTuples(tuple *pb.SPOTuple, ctx string, ts *[]*pb.SPOTuple, vLow, vHigh int64) {
	_, _, exist := n3ic.SubExist(tuple, ctx, vLow, vHigh)
	if !exist {
		// fPln("subject does not exist !")
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

// GetObjVer : we assume the return is unique, so use "fast" way to get the result
func (n3ic *DBClient) GetObjVer(tuple *pb.SPOTuple, ctx string) (string, int64) {
	// *** slow, but can get the last one ***
	// if obj, ver, found := n3ic.GetObj(tuple, 0, ctx, 0, 0); found {
	// 	return obj, ver
	// }
	// return "", -1

	if _, _, objs, vers, found := n3ic.GetObjs(tuple, ctx, false, false, 0, 0); found {
		return objs[0], vers[0]
	}
	return "", -1
}

// GetObj : (return object, version, IsFound)
// func (n3ic *DBClient) GetObj(tuple *pb.SPOTuple, offset int, ctx string, vLow, vHigh int64) (obj string, ver int64, found bool) {

// 	ver = -1
// 	s, p := tuple.Subject, tuple.Predicate
// 	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
// 	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

// 	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
// 	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' `+vChkL+vChkH, s, p)
// 	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1 OFFSET %d`, orderByTm, offset)

// 	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
// 	PE(e, resp.Error())
// 	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
// 		// pln(resp.Results[0].Series[0].Values)
// 		// pln(resp.Results[0].Series[0].Values[0])
// 		// pln(resp.Results[0].Series[0].Values[0][1]) /* [0] is time, [1] is object, as SELECT ... */

// 		lastItem := resp.Results[0].Series[0].Values[0]
// 		obj = lastItem[1].(string)
// 		ver = Must(lastItem[2].(json.Number).Int64()).(int64)
// 		found = true
// 	}
// 	return
// }

// GetSubStruct :
// func (n3ic *DBClient) getSubStruct(tuple *pb.SPOTuple, ctx string, vLow, vHigh int64) (obj string, ver int64, found bool) {
// 	// pln("looking for sub struct ...")

// 	ver = -1
// 	s, p := tuple.Predicate, "::"
// 	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
// 	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

// 	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
// 	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' AND version!=0 `+vChkL+vChkH, s, p)
// 	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)

// 	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
// 	PE(e, resp.Error())
// 	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
// 		lastItem := resp.Results[0].Series[0].Values[0]
// 		obj = lastItem[1].(string)
// 		ver = Must(lastItem[2].(json.Number).Int64()).(int64)
// 		found = true
// 	}
// 	return
// }

// GetArrInfo :
// func (n3ic *DBClient) getArrInfo(tuple *pb.SPOTuple, ctx string, vLow, vHigh int64) (cnt int, ver int64, found bool) {
// 	// pln("looking for array info ...")

// 	ver = -1
// 	s, p := tuple.Predicate, tuple.Subject
// 	vChkL := u.TerOp(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
// 	vChkH := u.TerOp(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

// 	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
// 	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' AND version!=0 `+vChkL+vChkH, s, p)
// 	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)

// 	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
// 	PE(e, resp.Error())
// 	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
// 		lastItem := resp.Results[0].Series[0].Values[0]
// 		cnt = u.Str(lastItem[1].(string)).ToInt()
// 		ver = Must(lastItem[2].(json.Number).Int64()).(int64)
// 		found = true
// 	}
// 	return
// }

// QueryTuple :
// func (n3ic *DBClient) QueryTuple(tuple *pb.SPOTuple, offset int, revArr bool, ctx, pathDel, childDel string,
// 	ts *[]*pb.SPOTuple, arrInfo *map[string]int, vLow, vHigh int64) {

// 	_, _, exist := n3ic.SubExist(tuple, ctx, vLow, vHigh)
// 	if !exist {
// 		// fPln("subject does not exist !")
// 		return
// 	}

// 	if obj, ver, ok := n3ic.GetObj(tuple, offset, ctx, vLow, vHigh); ok {
// 		// fPf("got object .................. %d \n", ver)
// 		tuple.Object = obj
// 		*ts = append(*ts, &pb.SPOTuple{
// 			Subject:   tuple.Subject,
// 			Predicate: tuple.Predicate,
// 			Object:    obj,
// 			Version:   ver,
// 		})
// 		return
// 	}

// 	if stru, _, ok := n3ic.getSubStruct(tuple, ctx, 0, 0); ok {
// 		// fPf("got sub-struct, more work ------> %s\n", stru)

// 		/* Array Element */
// 		if i := sI(stru, "[]"); i == 0 {
// 			nArr, _, _ := n3ic.getArrInfo(tuple, ctx, 0, 0)
// 			// pln(nArr)
// 			tuple.Predicate += fSf("%s%s", pathDel, stru[i+2:])
// 			// pln(tuple.Predicate)
// 			(*arrInfo)[tuple.Predicate] = nArr /* keep the array */

// 			for k := 0; k < nArr; k++ {
// 				n3ic.QueryTuple(tuple, u.TerOp(revArr, nArr-k-1, k).(int), revArr, ctx, pathDel, childDel, ts, arrInfo, vLow, vHigh)
// 			}
// 			tuple.Predicate = u.Str(tuple.Predicate).RmTailFromLast(pathDel)
// 			return
// 		}

// 		/* None Array Element */
// 		subs := strings.Split(stru, childDel)
// 		for _, s := range subs {
// 			tuple.Predicate += fSf("%s%s", pathDel, s)
// 			// pln(tuple.Predicate)
// 			n3ic.QueryTuple(tuple, offset, revArr, ctx, pathDel, childDel, ts, arrInfo, vLow, vHigh)
// 			tuple.Predicate = u.Str(tuple.Predicate).RmTailFromLast(pathDel)
// 		}
// 		return

// 		/* Only for debugging */
// 		// tuple.Object = stru
// 		// *ts = append(*ts, tuple)
// 		// return
// 	}
// }

// ************************************************************************************

// type arrOptTupleInfo struct {
// 	nArray   int     /* 'this' tuple's array's count */
// 	nTuple   int     /* 'this' tuple count in real */
// 	indices  []int   /* optional tuple's index in ts */
// 	versions []int64 /* optional tuple's version */
// }

// func getVersBySPO(ts *[]*pb.SPOTuple, spo ...string) (vers []int64, indices []int) {
// 	s, p, o := spo[0], spo[1], spo[2]
// 	for i, t := range *ts {
// 		if (s != "" && p != "" && o != "") && (t.Subject == s && t.Predicate == p && t.Object == o) {
// 			goto APPEND
// 		} else if (s != "" && p != "" && o == "") && (t.Subject == s && t.Predicate == p) {
// 			goto APPEND
// 		} else if (s != "" && p == "" && o != "") && (t.Subject == s && t.Object == o) {
// 			goto APPEND
// 		} else if (s == "" && p != "" && o != "") && (t.Predicate == p && t.Object == o) {
// 			goto APPEND
// 		} else if (s != "" && p == "" && o == "") && (t.Subject == s) {
// 			goto APPEND
// 		} else if (s == "" && p != "" && o == "") && (t.Predicate == p) {
// 			goto APPEND
// 		} else if (s == "" && p == "" && o != "") && (t.Object == o) {
// 			goto APPEND
// 		} else {
// 			continue
// 		}
// 	APPEND:
// 		vers = append(vers, t.Version)
// 		indices = append(indices, i)
// 	}
// 	return
// }

// // AdjustOptionalTuples : according to tuple version, adjust some tuples' order in sub array tuple
// func (n3ic *DBClient) AdjustOptionalTuples(ts *[]*pb.SPOTuple, arrInfo *map[string]int) {
// 	mapTuplesInArr := make(map[string]*arrOptTupleInfo)
// 	for ituple, tuple := range *ts {
// 		p, v := tuple.Predicate, tuple.Version
// 		if ok, nArr := u.Str(p).CoverAnyKeyInMapSI(*arrInfo); ok {
// 			// fPf("%d : %s : %d # %d\n", ituple, p, nArr, v)
// 			if _, ok := mapTuplesInArr[p]; !ok {
// 				mapTuplesInArr[p] = &arrOptTupleInfo{}
// 			}
// 			mapTuplesInArr[p].nArray = nArr
// 			mapTuplesInArr[p].nTuple++
// 			mapTuplesInArr[p].indices = append(mapTuplesInArr[p].indices, ituple)
// 			mapTuplesInArr[p].versions = append(mapTuplesInArr[p].versions, v)
// 		}
// 	}

// 	mapOptTuple := make(map[string]*arrOptTupleInfo)
// 	for k, v := range mapTuplesInArr {
// 		if v.nArray != v.nTuple {
// 			mapOptTuple[k] = v
// 		}
// 	}

// 	optTuples, aboveTuples := []*pb.SPOTuple{}, []*pb.SPOTuple{}
// 	for _, v := range mapOptTuple {
// 		for i, tIdx := range v.indices {
// 			this, above := (*ts)[tIdx], (*ts)[tIdx-1]
// 			// fPf("%d --- %v\n", tIdx, this)
// 			if vers, indicesTS := getVersBySPO(ts, "", above.Predicate, ""); len(vers) > 0 {
// 				tVer := v.versions[i]
// 				_, idx := u.I64(tVer).Nearest(vers...)
// 				posShouldAfter := indicesTS[idx]
// 				shouldAfter := (*ts)[posShouldAfter]
// 				// fPf("Should be after : %d --- %v\n", posShouldBeAfter, shouldAfter)
// 				optTuples, aboveTuples = append(optTuples, this), append(aboveTuples, shouldAfter)
// 			}
// 		}
// 	}

// 	/******************************************/

// 	tempts := make([]interface{}, len(*ts))
// 	for j, t := range *ts {
// 		tempts[j] = t
// 	}
// 	for i := 0; i < len(optTuples); i++ {
// 		tempts, _, _, _ = u.GArr(tempts).MoveItemAfter(
// 			func(move interface{}) bool { return move == optTuples[i] },
// 			func(after interface{}) bool { return after == aboveTuples[i] },
// 		)
// 	}
// 	for k, t := range tempts {
// 		(*ts)[k] = t.(*pb.SPOTuple)
// 	}
// }
