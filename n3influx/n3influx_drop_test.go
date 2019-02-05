package n3influx

import "testing"

func TestDropCtx(t *testing.T) {
	defer func() { uPH(recover(), "./log.txt", true) }()
	dbClient, e := NewPublisher()
	uPE(e)

	dbClient.DropCtx("D3E34F41-9D75-101A-8C3D-00AA001A1652")
}
