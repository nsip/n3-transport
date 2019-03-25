package n3node

import (
	"fmt"
	"strings"

	u "github.com/cdutwhu/go-util"
	"golang.org/x/sync/syncmap"
)

var (
	PE   = u.PanicOnError
	PE1  = u.PanicOnError1
	PH   = u.PanicHandle
	PC   = u.PanicOnCondition
	Must = u.Must

	fPln = fmt.Println
	fPf  = fmt.Printf
	fSf  = fmt.Sprintf

	sHP  = strings.HasPrefix
	sHS  = strings.HasSuffix
	sSpl = strings.Split

	prevID        = ""
	prevPred      = ""
	prevVer       int64
	startVer      int64
	verMeta       int64 = 1
	mapIDVQueue         = make(map[string][]int64)
	mapVerInDBChk       = make(map[string]int64)
	mapTickets          = syncmap.Map{}
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
)

const (
	DEADMARK      = "TOMBSTONE"
	TERMMARK      = "ENDENDEND"
	DELAY_CONTEST = 500
	DELAY_CHKTERM = 30
)
