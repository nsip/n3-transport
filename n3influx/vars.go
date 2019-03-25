package n3influx

import (
	"fmt"
	"strings"

	u "github.com/cdutwhu/go-util"
)

var (
	PE   = u.PanicOnError
	PE1  = u.PanicOnError1
	PH   = u.PanicHandle
	PC   = u.PanicOnCondition
	Must = u.Must

	fPf  = fmt.Printf
	fSf  = fmt.Sprintf
	fPln = fmt.Println

	sI   = strings.Index
	sC   = strings.Contains
	sSpl = strings.Split
)

const (
	db         = "tuples"
	orderByVer = "version" /* NOT supported */
	orderByTm  = "time"    /* only ORDER BY time supported at this time */
)
