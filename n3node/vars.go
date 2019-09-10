package n3node

import (
	"fmt"
	"strings"
	"sync"

	u "github.com/cdutwhu/go-util"
	w "github.com/cdutwhu/go-wrappers"
)

var (
	must        = u.Must
	ph          = u.PH
	IF          = u.IF
	matchAssign = u.MatchAssign
	trueAssign  = u.TrueAssign
	GoFn        = u.GoFn

	INum2Str  = w.INum2Str
	IArrEleIn = w.IArrEleIn
	IArrRmRep = w.IArrRmRep

	fPln = fmt.Println
	fPf  = fmt.Printf
	fSf  = fmt.Sprintf

	sSpl = strings.Split

	prevIDv  = ""
	prevVerV int64
	prevIDs  = ""
	prevVerS int64
	prevIDa  = ""
	prevVerA int64

	startVer  int64
	mIDvQueue = make(map[string][]int64)
	mIDsQueue = make(map[string][]int64)
	mIDaQueue = make(map[string][]int64)
	// mTickets  = syncmap.Map{}

	//                        object     context    path   R/W
	pcObjCtxPathRW = make(map[string]map[string]map[string]string)
	forroot        = ""
	forctx         = ""

	mpObjIDVer *sync.Map
)

type (
	metaData struct {
		ID       string
		StartVer int64
		EndVer   int64
		Ver      int64
	}

	ticket struct {
		tktID string
		idx   string
	}

	S   = w.Str
	Ss  = w.Strs
	I64 = w.I64
)

const (
	MARKDead      = "TOMBSTONE"
	MARKTerm      = "--------------------------------------" // len(uuid) + 2 : 38
	DELIPath      = " ~ "
	DELIChild     = " + "
	MARKDelID     = "00000000-0000-0000-0000-000000000000"
	DELAY_CONTEST = 2000
	DELAY_CHKTERM = 5000
	NOREAD        = "******"
	NOWRITE       = "------"
	BIGVER        = 999999999
)
