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
		ignrFld := ignoreFields[0]
		switch {
		case IArrEleIn(ignrFld, SINDList):
			qWhere = fSf(`WHERE predicate='%s' AND object='%s' AND version=%d `, p, o, v)
		case IArrEleIn(ignrFld, PINDList):
			qWhere = fSf(`WHERE subject='%s' AND object='%s' AND version=%d `, s, o, v)
		case IArrEleIn(ignrFld, OINDList):
			qWhere = fSf(`WHERE subject='%s' AND predicate='%s' AND version=%d `, s, p, v)
		case IArrEleIn(ignrFld, VINDList):
			qWhere = fSf(`WHERE subject='%s' AND predicate='%s' AND object='%s' `, s, p, o)
		}
	}

	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	return len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0
}

// PairOfSPOExists : spoIND1 -> "s", "p"; spoIND2 -> "p", "o". if exists, get the last one
func (n3ic *DBClient) PairOfSPOExists(ctx, spo1, spo2, spoIND1, spoIND2 string, vLow, vHigh int64) (r string, exist bool) {

	vChkL := IF(vLow > 0, fSf(" AND version>=%d ", vLow), "").(string)
	vChkH := IF(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect, qWhere := "", ""
	switch {
	case IArrEleIn(spoIND1, SINDList) && IArrEleIn(spoIND2, PINDList):
		s, p := spo1, spo2
		qSelect = fSf(`SELECT version, object FROM "%s" `, ctx)
		qWhere = fSf(`WHERE subject='%s' AND predicate='%s' `+vChkL+vChkH, s, p)

	case IArrEleIn(spoIND1, SINDList) && IArrEleIn(spoIND2, OINDList):
		s, o := spo1, spo2
		qSelect = fSf(`SELECT version, predicate FROM "%s" `, ctx)
		qWhere = fSf(`WHERE subject='%s' AND object='%s' `+vChkL+vChkH, s, o)

	case IArrEleIn(spoIND1, PINDList) && IArrEleIn(spoIND2, OINDList):
		p, o := spo1, spo2
		qSelect = fSf(`SELECT version, subject FROM "%s" `, ctx)
		qWhere = fSf(`WHERE predicate='%s' AND object='%s' `+vChkL+vChkH, p, o)
	}

	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())

	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		lastItem := resp.Results[0].Series[0].Values[0]
		r, exist = lastItem[2].(string), true
	}
	return
}

// OneOfSPOExists : if exists, get the last one
func (n3ic *DBClient) OneOfSPOExists(ctx, spo, spoIND string, vLow, vHigh int64) (r1, r2 string, exist bool) {

	vChkL := IF(vLow > 0, fSf(" AND version>=%d ", vLow), "").(string)
	vChkH := IF(vHigh > 0, fSf(" AND version<=%d ", vHigh), "").(string)

	qSelect, qWhere := "", ""
	switch {
	case IArrEleIn(spoIND, SINDList):
		qSelect = fSf(`SELECT version, predicate, object FROM "%s" `, ctx)
		qWhere = fSf(`WHERE subject='%s' `+vChkL+vChkH, spo)
	case IArrEleIn(spoIND, PINDList):
		qSelect = fSf(`SELECT version, subject, object FROM "%s" `, ctx)
		qWhere = fSf(`WHERE predicate='%s' `+vChkL+vChkH, spo)
	case IArrEleIn(spoIND, OINDList):
		qSelect = fSf(`SELECT version, subject, predicate FROM "%s" `, ctx)
		qWhere = fSf(`WHERE object='%s' `+vChkL+vChkH, spo)
	}

	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())

	// pln(resp.Results[0].Series[0].Values[0][1]) /* [0] is time, [1] is as SELECT ... */
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		lastItem := resp.Results[0].Series[0].Values[0]
		r1, r2, exist = lastItem[2].(string), lastItem[3].(string), true
	}
	return
}

// RootByID :
func (n3ic *DBClient) RootByID(objID, ctx, del string) string {
	if !S(ctx).HS("-meta") && S(objID).IsUUID() {
		if p, _, ok := n3ic.OneOfSPOExists(ctx, objID, "S", -1, -1); ok && S(p).Contains(del) {
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

// SingleListOfSPO :
func (n3ic *DBClient) SingleListOfSPO(ctx, spoIND string) (rst []string) {
	wanted := "subject"
	switch {
	case IArrEleIn(spoIND, SINDList):
		wanted = "subject"
	case IArrEleIn(spoIND, PINDList):
		wanted = "predicate"
	case IArrEleIn(spoIND, OINDList):
		wanted = "object"
	}
	qStr := fSf(`SELECT distinct(%s) from (SELECT version, %s FROM "%s")`, wanted, wanted, ctx)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {
			rst = append(rst, l[1].(string))
		}
	}
	return
}

// PairListOfSPO :
func (n3ic *DBClient) PairListOfSPO(ctx, spoExcl string) (rst1, rst2, rstExcl []string) {

	r1, r2 := []string{}, []string{}
	spoIND1, spoIND2 := "", ""

	switch {
	case IArrEleIn(spoExcl, SINDList):
		r1, r2 = n3ic.SingleListOfSPO(ctx, "P"), n3ic.SingleListOfSPO(ctx, "O")
		spoIND1, spoIND2 = "P", "O"
	case IArrEleIn(spoExcl, PINDList):
		r1, r2 = n3ic.SingleListOfSPO(ctx, "S"), n3ic.SingleListOfSPO(ctx, "O")
		spoIND1, spoIND2 = "S", "O"
	case IArrEleIn(spoExcl, OINDList):
		r1, r2 = n3ic.SingleListOfSPO(ctx, "S"), n3ic.SingleListOfSPO(ctx, "P")
		spoIND1, spoIND2 = "S", "P"
	}

	for _, i1 := range r1 {
		for _, i2 := range r2 {
			if r, ok := n3ic.PairOfSPOExists(ctx, i1, i2, spoIND1, spoIND2, -1, -1); ok {
				rst1, rst2, rstExcl = append(rst1, i1), append(rst2, i2), append(rstExcl, r)
			}
		}
	}

	return
}

// LastPOByS :
func (n3ic *DBClient) LastPOByS(ctx, sub string) (pred, obj string) {
	qSelect := fSf(`SELECT predicate, object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' `, sub)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {
			return l[1].(string), l[2].(string)
		}
	}
	return
}

// POsByS :
func (n3ic *DBClient) POsByS(ctx, sub, predExcl, objExcl string) (preds, objs []string) {
	qSelect := fSf(`SELECT predicate, object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' `, sub)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {
			p, o := l[1].(string), l[2].(string)
			if p != predExcl && o != objExcl {
				preds = append(preds, p)
				objs = append(objs, o)
			}
		}
	}
	return
}

// LastOBySP :
func (n3ic *DBClient) LastOBySP(ctx, sub, pred string) string {
	qSelect := fSf(`SELECT object, version FROM "%s" `, ctx)
	qWhere := fSf(`WHERE subject='%s' AND predicate='%s' `, sub, pred)
	qStr := qSelect + qWhere + fSf(`ORDER BY %s DESC LIMIT 1`, orderByTm)
	resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
	pe(e, resp.Error())
	if len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, l := range resp.Results[0].Series[0].Values {
			return l[1].(string)
		}
	}
	return ""
}

// OsBySP : (return objects, versions, IsFound)
func (n3ic *DBClient) OsBySP(tuple *pb.SPOTuple, ctx string, extSub, extPred bool, vLow, vHigh int64) (subs, preds, objs []string, vers []int64, found bool) {

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

// TuplesBySP :
func (n3ic *DBClient) TuplesBySP(tuple *pb.SPOTuple, ctx string, ts *[]*pb.SPOTuple, vLow, vHigh int64) {
	_, _, exist := n3ic.OneOfSPOExists(ctx, tuple.Subject, "S", vLow, vHigh)
	if !exist {
		// fPln("subject does not exist !")
		return
	}
	if subs, preds, objs, vers, ok := n3ic.OsBySP(tuple, ctx, false, true, vLow, vHigh); ok {
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

// LastOV : we assume the return is unique, so use "fast" way to get the result
func (n3ic *DBClient) LastOV(tuple *pb.SPOTuple, ctx string) (string, int64) {
	if _, _, objs, vers, found := n3ic.OsBySP(tuple, ctx, false, false, 0, 0); found {
		return objs[0], vers[0]
	}
	return "", -1
}

// Status :
func (n3ic *DBClient) Status(ObjID, ctx string) (exist, alive bool) {
	ctx = S(ctx).MkSuffix("-meta").V()
	ObjID = S(ObjID).RmPrefix("::").V()
	ObjID = S(ObjID).RmPrefix("[]").V()
	pred, _, exist := n3ic.OneOfSPOExists(ctx, ObjID, "S", -1, -1)
	return exist, pred != MARKDead
}

// ObjectCount :
func (n3ic *DBClient) ObjectCount(ctx, objIDIND string) int64 {

	qSelect := fSf(`SELECT count(*) FROM "%s" `, ctx)
	qWhere := fSf(`WHERE predicate=~/ ~ %s$/ AND predicate!~/ ~ [A-Za-z]+ ~ %s$/`, objIDIND, objIDIND)
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
