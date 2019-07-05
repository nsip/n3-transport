package n3influx

import (
	"encoding/json"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/nsip/n3-messages/messages/pb"
)

// TupleExists :
func (n3ic *DBClient) TupleExists(tuple *pb.SPOTuple, ctx string, ignoreFields ...string) bool {
	s, p, o, v := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version

	qSelect := fSf(`SELECT version, subject, predicate, object FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' AND object='%s' AND version=%d `, s, p, o, v)

	if L := len(ignoreFields); L == 1 {
		// PC(L != 0 && L != 1, fEf("Currently only support ignoring 1 Field(s)"))
		switch ignoreFields[0] {
		case "Subject", "subject", "Sub", "sub", "S", "s":
			qWhere = fSf(`WHERE predicate='%s' AND object='%s' AND version=%d `, p, o, v)
		case "Predicate", "predicate", "P", "p":
			qWhere = fSf(`WHERE subject='%s' AND object='%s' AND version=%d `, s, o, v)
		case "Object", "object", "Obj", "obj", "O", "o":
			qWhere = fSf(`WHERE subject='%s' AND predicate='%s' AND version=%d `, s, p, v)
		case "Version", "version", "Ver", "ver", "V", "v":
			qWhere = fSf(`WHERE subject='%s' AND predicate='%s' AND object='%s' `, s, p, o)
		}
	}

	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	return len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0
}

// SubjectExist :
func (n3ic *DBClient) SubjectExist(subject, ctx string, vLow, vHigh int64) (pred, obj string, exist bool) {

	vChkL := IF(vLow > 0, fSf(" AND version>=%d ", vLow), "").(string)
	vChkH := IF(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect := fSf(`SELECT version, predicate, object FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' `+vChkL+vChkH, subject)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())

	// pln(resp.Results[0].Series[0].Values[0][1]) /* [0] is time, [1] is as SELECT ... */
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		lastItem := resp.Results[0].Series[0].Values[0]
		pred, obj, exist = lastItem[2].(string), lastItem[3].(string), true
	}
	return
}

// RootByID :
func (n3ic *DBClient) RootByID(objID, ctx, del string) string {
	if !S(ctx).HS("-meta") && S(objID).IsUUID() {
		if p, _, ok := n3ic.SubjectExist(objID, ctx, 0, 0); ok && S(p).Contains(del) {
			return sSpl(p, del)[0]			
		}
	}
	return ""
}

// IDListByPathValue :
func (n3ic *DBClient) IDListByPathValue(tuple *pb.SPOTuple, ctx string, caseSensitive bool) (ids []string) {
	qSelect := fSf(`SELECT version, subject FROM "%s" `, ctx)
	qWhere := fSf(`WHERE predicate='%s' AND object='%s' `, tuple.Predicate, tuple.Object)
	if !caseSensitive {
		objRegex := regex4CaseIns(tuple.Object)
		qWhere = fSf(`WHERE predicate='%s' AND object=~/%s/`, tuple.Predicate, objRegex)
	}
	qStr := qSelect + qWhere + fSf(`ORDER BY %s`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	if len(resp.Results[0].Series) > 0 {
		for _, v := range resp.Results[0].Series[0].Values {
			id := v[2].(string)
			if S(id).IsUUID() {
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

// GetObjsBySP : (return objects, versions, IsFound)
func (n3ic *DBClient) GetObjsBySP(tuple *pb.SPOTuple, ctx string, extSub, extPred bool, vLow, vHigh int64) (subs, preds, objs []string, vers []int64, found bool) {

	s, p := tuple.Subject, tuple.Predicate
	vChkL := IF(vLow > 0, fSf(" AND version>=%d ", vLow), " AND version>0 ").(string)
	vChkH := IF(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

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
	pe(e, resp.Error())

	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {
			sub, pred, obj := l[1].(string), l[2].(string), l[3].(string)
			subs, preds, objs = append(subs, sub), append(preds, pred), append(objs, obj)
			vers = append(vers, must(l[4].(json.Number).Int64()).(int64))
			// fPln(pred, obj, ver)
		}
		found = true
	}
	return
}

// QueryTuplesBySP :
func (n3ic *DBClient) QueryTuplesBySP(tuple *pb.SPOTuple, ctx string, ts *[]*pb.SPOTuple, vLow, vHigh int64) {
	_, _, exist := n3ic.SubjectExist(tuple.Subject, ctx, vLow, vHigh)
	if !exist {
		// fPln("subject does not exist !")
		return
	}
	if subs, preds, objs, vers, ok := n3ic.GetObjsBySP(tuple, ctx, false, true, vLow, vHigh); ok {
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

// LastObjVer : we assume the return is unique, so use "fast" way to get the result
func (n3ic *DBClient) LastObjVer(tuple *pb.SPOTuple, ctx string) (string, int64) {
	if _, _, objs, vers, found := n3ic.GetObjsBySP(tuple, ctx, false, false, 0, 0); found {
		return objs[0], vers[0]
	}
	return "", -1
}

// Status :
func (n3ic *DBClient) Status(ObjID, ctx string) (exist, alive bool) {
	ctx = S(ctx).MkSuffix("-meta").V()
	ObjID = S(ObjID).RmPrefix("::").V()
	ObjID = S(ObjID).RmPrefix("[]").V()
	pred, _, exist := n3ic.SubjectExist(ObjID, ctx, -1, -1)
	return exist, pred != MARKDead
}

// ObjectCount :
func (n3ic *DBClient) ObjectCount(ctx, objIDMark string) int64 {

	qSelect := fSf(`SELECT count(*) FROM "%s" `, ctx)
	qWhere := fSf(`WHERE predicate=~/ ~ %s$/ AND predicate!~/ ~ [A-Za-z]+ ~ %s$/`, objIDMark, objIDMark)
	qStr := qSelect + qWhere
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	if len(resp.Results[0].Series) > 0 {
		for _, v := range resp.Results[0].Series[0].Values {
			// fPln("object count:", v[1])
			return must(v[1].(json.Number).Int64()).(int64)
		}
	}
	return 0
}
