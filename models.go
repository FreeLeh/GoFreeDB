package freeleh

import (
	"errors"
)

type KVMode int

const (
	KVModeDefault    KVMode = 0
	KVModeAppendOnly KVMode = 1

	scratchpadBooked          = "BOOKED"
	scratchpadSheetNameSuffix = "_scratch"
	defaultTableRange         = "A1:C5000000"
	defaultKeyColRange        = "A1:A5000000"

	getAppendQueryTemplate      = "=VLOOKUP(\"%s\", SORT(%s, 3, FALSE), 2, FALSE)"
	getDefaultQueryTemplate     = "=VLOOKUP(\"%s\", %s, 2, FALSE)"
	findKeyA1RangeQueryTemplate = "=MATCH(\"%s\", %s, 0)"

	naValue = "#N/A"
)

var ErrKeyNotFound = errors.New("error key not found")
