package lark

import (
	"context"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type appendMode string
type valueRenderOption string

const (
	appendModeInsert    appendMode = "INSERT_ROWS"
	appendModeOverwrite appendMode = "OVERWRITE"

	valueRenderOptionFormattedValue = "FormattedValue"

	// Must skip the header by using the A1 range (start from row 2).
	// Otherwise, the query row might return the header on Get Single Range.
	queryFormulaTemplate = "=QUERY('%s'!A2:Z500000, \"%s\", 0)"

	createSheetURL     = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/sheets_batch_update"
	deleteSheetsURL    = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/sheets_batch_update"
	appendValuesURL    = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/values_append?insertDataOption=%s"
	getSheetsURL       = "https://open.larksuite.com/open-apis/sheets/v3/spreadsheets/%s/sheets/query"
	batchUpdateRowsURL = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/values_batch_update"
	getSingleRangeURL  = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/values/%s?valueRenderOption=%s"

	apiStatusCodeOK = 0

	// Currently limited to 26.
	// Otherwise, the sheet creation must extend the column as well to make the rowGetIndicesQueryTemplate formula works.
	// TODO(edocsss): add an option to extend the number of columns.
	maxColumn = 26

	scratchpadBooked            = "BOOKED"
	scratchpadSheetNameTemplate = "%s_scratch:%d"

	// Unlike Google Sheets, for Lark Sheets, each scratchpad only has one user.
	// Hence, the range doesn't really have to be so big.
	defaultScratchpadTableRange = "A1:Z1000"

	selectStmtQueryRangeTemplate = "A:%s"

	rowIdxCol     = "_rid"
	rowIdxFormula = "=ROW()"

	invalidValue = "#VALUE!"
	naValue      = "#N/A"
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

//go:generate mockgen -destination=models_mock.go -package=lark -build_constraint=gofreedb_test . sheetsWrapper
type sheetsWrapper interface {
	CreateSheet(
		ctx context.Context,
		spreadsheetToken string,
		sheetName string,
	) error

	GetSheets(
		ctx context.Context,
		spreadsheetToken string,
	) (GetSheetsResult, error)

	DeleteSheets(
		ctx context.Context,
		spreadsheetToken string,
		sheetIDs []string,
	) error

	QueryRows(
		ctx context.Context,
		spreadsheetToken string,
		sheetName string,
		a1Range models.A1Range,
		query string,
	) (QueryRowsResult, error)

	OverwriteRows(
		ctx context.Context,
		spreadsheetToken string,
		a1Range models.A1Range,
		values [][]interface{},
	) (InsertRowsResult, error)

	BatchUpdateRows(
		ctx context.Context,
		spreadsheetToken string,
		requests []BatchUpdateRowsRequest,
	) ([]BatchUpdateRowsResult, error)

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

type baseHTTPResp[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type insertRowsHTTPResp struct {
	Updates struct {
		UpdatedRange   string `json:"updatedRange"`
		UpdatedRows    int64  `json:"updatedRows"`
		UpdatedColumns int64  `json:"updatedColumns"`
		UpdatedCells   int64  `json:"updatedCells"`
	} `json:"updates"`
}

type InsertRowsResult struct {
	UpdatedRange   models.A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
}

type getSheetsHTTPResp struct {
	Sheets []Sheet `json:"sheets"`
}

type GetSheetsResult struct {
	Sheets []Sheet `json:"sheets"`
}

type Sheet struct {
	SheetID string `json:"sheet_id"`
	Title   string `json:"title"`
}

type BatchUpdateRowsRequest struct {
	A1Range models.A1Range
	Values  [][]interface{}
}

type batchUpdateRowsHTTPResp struct {
	Responses []struct {
		UpdatedRange   string `json:"updatedRange"`
		UpdatedRows    int64  `json:"updatedRows"`
		UpdatedColumns int64  `json:"updatedColumns"`
		UpdatedCells   int64  `json:"updatedCells"`
	} `json:"responses"`
}

type BatchUpdateRowsResult struct {
	UpdatedRange   models.A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
}

type getSingleRangeHTTPResp struct {
	ValueRange struct {
		MajorDimension string          `json:"majorDimension"`
		Range          string          `json:"range"`
		Values         [][]interface{} `json:"values"`
	} `json:"valueRange"`
}

type getSingleRangeResult struct {
	MajorDimension string
	Range          string
	Values         [][]interface{}
}

type QueryRowsResult struct {
	Rows [][]interface{}
}
