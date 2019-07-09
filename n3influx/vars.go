package n3influx

import (
	"fmt"
	"strings"

	u "github.com/cdutwhu/go-util"
	w "github.com/cdutwhu/go-wrappers"
)

type (
	S  = w.Str
	Ss = w.Strs
)

var (
	pe   = u.PanicOnError
	pe1  = u.PanicOnError1
	ph   = u.PanicHandle
	pc   = u.PanicOnCondition
	must = u.Must
	IF   = u.IF

	IArrIntersect = w.IArrIntersect

	fPf  = fmt.Printf
	fSf  = fmt.Sprintf
	fPln = fmt.Println
	fEf  = fmt.Errorf

	sSpl = strings.Split
)

const (
	db         = "tuples"
	orderByVer = "version" /* NOT supported */
	orderByTm  = "time"    /* only ORDER BY time supported at this time */
	MARKDead   = "TOMBSTONE"
	MARKDelID  = "00000000-0000-0000-0000-000000000000"
)
