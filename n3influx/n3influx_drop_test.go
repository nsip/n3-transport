package n3influx

import "testing"

func TestDropCtx(t *testing.T) {
	defer func() { ph(recover(), "./log.txt") }()
	dbClient := must(NewDBClient()).(*DBClient)
	dbClient.DropCtx("D3E34F41-9D75-101A-8C3D-00AA001A1652")
}
