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
	IArrEleIn     = w.IArrEleIn

	fPf  = fmt.Printf
	fSf  = fmt.Sprintf
	fPln = fmt.Println
	fEf  = fmt.Errorf

	sSpl = strings.Split

	SINDList = Ss([]string{"subject", "sub", "s", "SUBJECT", "SUB", "S"})
	PINDList = Ss([]string{"predicate", "pred", "p", "PREDICATE", "PRED", "P"})
	OINDList = Ss([]string{"object", "obj", "o", "OBJECT", "OBJ", "O"})
	VINDList = Ss([]string{"version", "ver", "v", "VERSION", "VER", "V"})
)

const (
	db         = "tuples"
	orderByVer = "version" /* time NOT supported */
	orderByTm  = "time"    /* only ORDER BY time supported at this time */
	MARKDead   = "TOMBSTONE"
	MARKDelID  = "00000000-0000-0000-0000-000000000000"
)
