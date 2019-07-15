package n3node

import (
	"testing"

	db "../n3influx"
)

func TestMKPrivCtrl(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(db.NewDBClient()).(*db.DBClient)
	rst := mkPrivCtrl(dbClient)
	for k, v := range rst {
		fPln(k, v)
	}
}
