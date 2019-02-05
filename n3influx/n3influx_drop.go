package n3influx

import (
	influx "github.com/influxdata/influxdb1-client/v2"
)

// DropCtx :
func (n3ic *DBClient) DropCtx(ctxNames ...string) {
	// pln("dropping table ...")
	for _, ctx := range ctxNames {
		qStr := fSf(`DROP MEASUREMENT "%s"`, ctx)
		resp, e := n3ic.cl.Query(influx.NewQuery(qStr, db, ""))
		PE(e, resp.Error())
	}
}
