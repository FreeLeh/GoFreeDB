package freedb

import (
	"context"
	"errors"
	"regexp"

	"github.com/FreeLeh/GoFreeDB/google/auth"

	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
)

// KVMode defines the mode of the key value store.
// For more details, please read the README file.
type KVMode int

const (
	KVModeDefault    KVMode = 0
	KVModeAppendOnly KVMode = 1
)

const (
	// Currently limited to 26.
	// Otherwise, the sheet creation must extend the column as well to make the rowGetIndicesQueryTemplate formula works.
	// TODO(edocsss): add an option to extend the number of columns.
	maxColumn = 26

	scratchpadBooked          = "BOOKED"
	scratchpadSheetNameSuffix = "_scratch"

	defaultKVTableRange    = "A1:C5000000"
	defaultKVKeyColRange   = "A1:A5000000"
	defaultKVFirstRowRange = "A1:C1"

	kvGetAppendQueryTemplate      = "=VLOOKUP(\"%s\", SORT(%s, 3, FALSE), 2, FALSE)"
	kvGetDefaultQueryTemplate     = "=VLOOKUP(\"%s\", %s, 2, FALSE)"
	kvFindKeyA1RangeQueryTemplate = "=MATCH(\"%s\", %s, 0)"

	naValue       = "#N/A"
	errorValue    = "#ERROR!"
	rowIdxCol     = "_rid"
	rowIdxFormula = "=ROW()"
)

// ErrKeyNotFound is returned only for the key-value store and when the key does not exist.
var (
	ErrKeyNotFound = errors.New("error key not found")
)

// FreeDBGoogleAuthScopes specifies the list of Google Auth scopes required to run FreeDB implementations properly.
var (
	FreeDBGoogleAuthScopes = auth.GoogleSheetsReadWrite
)

var (
	defaultRowHeaderRange    = "A1:" + generateColumnName(maxColumn-1) + "1"
	defaultRowFullTableRange = "A2:" + generateColumnName(maxColumn-1)
	rowDeleteRangeTemplate   = "A%d:" + generateColumnName(maxColumn-1) + "%d"

	// The first condition `_rid IS NOT NULL` is necessary to ensure we are just updating rows that are non-empty.
	// This is required for UPDATE without WHERE clause (otherwise it will see every row as update target).
	rowWhereNonEmptyConditionTemplate = rowIdxCol + " is not null AND %s"
	rowWhereEmptyConditionTemplate    = rowIdxCol + " is not null"

	googleSheetSelectStmtStringKeyword = regexp.MustCompile("^(date|datetime|timeofday)")
)

// Codec is an interface for encoding and decoding the data provided by the client.
// At the moment, only key-value store requires data encoding.
type Codec interface {
	Encode(value []byte) (string, error)
	Decode(value string) ([]byte, error)
}

type sheetsWrapper interface {
	CreateSpreadsheet(ctx context.Context, title string) (string, error)
	GetSheetNameToID(ctx context.Context, spreadsheetID string) (map[string]int64, error)
	CreateSheet(ctx context.Context, spreadsheetID string, sheetName string) error
	DeleteSheets(ctx context.Context, spreadsheetID string, sheetIDs []int64) error
	InsertRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (sheets.InsertRowsResult, error)
	OverwriteRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (sheets.InsertRowsResult, error)
	UpdateRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (sheets.UpdateRowsResult, error)
	BatchUpdateRows(ctx context.Context, spreadsheetID string, requests []sheets.BatchUpdateRowsRequest) (sheets.BatchUpdateRowsResult, error)
	QueryRows(ctx context.Context, spreadsheetID string, sheetName string, query string, skipHeader bool) (sheets.QueryRowsResult, error)
	Clear(ctx context.Context, spreadsheetID string, ranges []string) ([]string, error)
}

type colIdx struct {
	name string
	idx  int
}

type colsMapping map[string]colIdx

func (m colsMapping) NameMap() map[string]string {
	result := make(map[string]string, 0)
	for col, val := range m {
		result[col] = val.name
	}
	return result
}

// OrderBy defines the type of column ordering used for GoogleSheetRowStore.Select().
type OrderBy string

const (
	OrderByAsc  OrderBy = "ASC"
	OrderByDesc OrderBy = "DESC"
)

// ColumnOrderBy defines what ordering is required for a particular column.
// This is used for GoogleSheetRowStore.Select().
type ColumnOrderBy struct {
	Column  string
	OrderBy OrderBy
}
