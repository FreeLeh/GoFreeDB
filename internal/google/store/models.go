package store

import (
	"context"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"regexp"

	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
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

	rowIdxCol     = "_rid"
	rowIdxFormula = "=ROW()"
)

var (
	defaultRowHeaderRange    = "A1:" + models.GenerateColumnName(maxColumn-1) + "1"
	defaultRowFullTableRange = "A2:" + models.GenerateColumnName(maxColumn-1)
	rowDeleteRangeTemplate   = "A%d:" + models.GenerateColumnName(maxColumn-1) + "%d"

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
	GetSheetNameToID(
		ctx context.Context,
		spreadsheetID string,
	) (map[string]int64, error)

	CreateSheet(
		ctx context.Context,
		spreadsheetID string,
		sheetName string,
	) error

	DeleteSheets(
		ctx context.Context,
		spreadsheetID string,
		sheetIDs []int64,
	) error

	InsertRows(
		ctx context.Context,
		spreadsheetID string,
		a1Range models.A1Range,
		values [][]interface{},
	) (sheets.InsertRowsResult, error)

	OverwriteRows(
		ctx context.Context,
		spreadsheetID string,
		a1Range models.A1Range,
		values [][]interface{},
	) (sheets.InsertRowsResult, error)

	UpdateRows(
		ctx context.Context,
		spreadsheetID string,
		a1Range models.A1Range,
		values [][]interface{},
	) (sheets.UpdateRowsResult, error)

	BatchUpdateRows(
		ctx context.Context,
		spreadsheetID string,
		requests []sheets.BatchUpdateRowsRequest,
	) (sheets.BatchUpdateRowsResult, error)

	QueryRows(
		ctx context.Context,
		spreadsheetID string,
		sheetName string,
		query string,
		skipHeader bool,
	) (sheets.QueryRowsResult, error)

	Clear(
		ctx context.Context,
		spreadsheetID string,
		ranges []models.A1Range,
	) ([]string, error)
}
