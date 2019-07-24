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
	pe              = u.PanicOnError
	pe1             = u.PanicOnError1
	ph              = u.PanicHandle
	pc              = u.PanicOnCondition
	must            = u.Must
	IF              = u.IF
	matchAssign     = u.MatchAssign
	conditionAssign = u.ConditionAssign

	IArrIntersect = w.IArrIntersect
	IArrEleIn     = w.IArrEleIn

	fPf  = fmt.Printf
	fSf  = fmt.Sprintf
	fPln = fmt.Println
	fEf  = fmt.Errorf

	sSpl = strings.Split

	SINDList = Ss{"subject", "sub", "s", "SUBJECT", "SUB", "S"}
	PINDList = Ss{"predicate", "pred", "p", "PREDICATE", "PRED", "P"}
	OINDList = Ss{"object", "obj", "o", "OBJECT", "OBJ", "O"}
	VINDList = Ss{"version", "ver", "v", "VERSION", "VER", "V"}
)

const (
	db         = "tuples"
	orderByVer = "version" /* time NOT supported */
	orderByTm  = "time"    /* only ORDER BY time supported at this time */
	MARKDead   = "TOMBSTONE"
	MARKDelID  = "00000000-0000-0000-0000-000000000000"
	MARKTerm   = "--------------------------------------" // len(uuid) + 2 : 38
	DELIPath   = " ~ "
	DELIChild  = " + "
)
