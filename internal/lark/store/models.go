package store

import (
	"context"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/lark/sheets"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

const (
	// Currently limited to 26.
	// Otherwise, the sheet creation must extend the column as well to make the rowGetIndicesQueryTemplate formula works.
	// TODO(edocsss): add an option to extend the number of columns.
	maxColumn = 26

	scratchpadBooked          = "BOOKED"
	scratchpadSheetNameSuffix = "_scratch"

	defaultScratchpadTableRange = "A1:C5000000"

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
)

// Codec is an interface for encoding and decoding the data provided by the client.
// At the moment, only key-value store requires data encoding.
type Codec interface {
	Encode(value []byte) (string, error)
	Decode(value string) ([]byte, error)
}

type sheetsWrapper interface {
	CreateSheet(
		ctx context.Context,
		spreadsheetToken string,
		sheetName string,
	) error

	GetSheets(
		ctx context.Context,
		spreadsheetToken string,
	) (sheets.GetSheetsResult, error)

	DeleteSheets(
		ctx context.Context,
		spreadsheetToken string,
		sheetIDs []string,
	) error

	QueryRows(
		ctx context.Context,
		spreadsheetToken string,
		a1Range models.A1Range,
		query string,
	) (sheets.QueryRowsResult, error)

	OverwriteRows(
		ctx context.Context,
		spreadsheetToken string,
		a1Range models.A1Range,
		values [][]interface{},
	) (sheets.InsertRowsResult, error)

	BatchUpdateRows(
		ctx context.Context,
		spreadsheetToken string,
		requests []sheets.BatchUpdateRowsRequest,
	) ([]sheets.BatchUpdateRowsResult, error)

	Clear(
		ctx context.Context,
		spreadsheetToken string,
		ranges []models.A1Range,
	) error
}

func ridWhereClauseInterceptor(where string) string {
	if where == "" {
		return rowWhereEmptyConditionTemplate
	}
	return fmt.Sprintf(rowWhereNonEmptyConditionTemplate, where)
}
