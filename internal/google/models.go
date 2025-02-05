package google

import (
	"context"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"regexp"
)

type appendMode string

const (
	majorDimensionRows                      = "ROWS"
	valueInputUserEntered                   = "USER_ENTERED"
	responseValueRenderFormatted            = "FORMATTED_VALUE"
	appendModeInsert             appendMode = "INSERT_ROWS"
	appendModeOverwrite          appendMode = "OVERWRITE"

	queryRowsURLTemplate = "https://docs.google.com/spreadsheets/d/%s/gviz/tq"
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
	) (InsertRowsResult, error)

	OverwriteRows(
		ctx context.Context,
		spreadsheetID string,
		a1Range models.A1Range,
		values [][]interface{},
	) (InsertRowsResult, error)

	UpdateRows(
		ctx context.Context,
		spreadsheetID string,
		a1Range models.A1Range,
		values [][]interface{},
	) (UpdateRowsResult, error)

	BatchUpdateRows(
		ctx context.Context,
		spreadsheetID string,
		requests []BatchUpdateRowsRequest,
	) (BatchUpdateRowsResult, error)

	QueryRows(
		ctx context.Context,
		spreadsheetID string,
		sheetName string,
		query string,
		skipHeader bool,
	) (QueryRowsResult, error)

	Clear(
		ctx context.Context,
		spreadsheetID string,
		ranges []models.A1Range,
	) ([]string, error)
}

type InsertRowsResult struct {
	UpdatedRange   models.A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
	InsertedValues [][]interface{}
}

type UpdateRowsResult struct {
	UpdatedRange   models.A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
	UpdatedValues  [][]interface{}
}

type BatchUpdateRowsRequest struct {
	A1Range models.A1Range
	Values  [][]interface{}
}

type BatchUpdateRowsResult []UpdateRowsResult

/*
	{
		"version":"0.6",
		"reqId":"0",
		"status":"ok",
		"sig":"141753603",
		"table":{
			"cols":[
				{"id":"A","label":"","type":"string"},
				{"id":"B","label":"","type":"number","pattern":"General"}
			],
			"rows":[
				{"c":[{"v":"k1"},{"v":103.0,"f":"103"}]},
				{"c":[{"v":"k2"},{"v":111.0,"f":"111"}]},
				{"c":[{"v":"k3"},{"v":123.0,"f":"123"}]}
			],
			"parsedNumHeaders":0
		}
	}
*/
type rawQueryRowsResult struct {
	Table rawQueryRowsResultTable `json:"table"`
}

func (r rawQueryRowsResult) toQueryRowsResult() (QueryRowsResult, error) {
	result := QueryRowsResult{
		Rows: make([][]interface{}, len(r.Table.Rows)),
	}

	for rowIdx, row := range r.Table.Rows {
		result.Rows[rowIdx] = make([]interface{}, len(row.Cells))
		for cellIdx, cell := range row.Cells {
			val, err := r.convertRawValue(cellIdx, cell)
			if err != nil {
				return QueryRowsResult{}, err
			}
			result.Rows[rowIdx][cellIdx] = val
		}
	}

	return result, nil
}

func (r rawQueryRowsResult) convertRawValue(cellIdx int, cell rawQueryRowsResultCell) (interface{}, error) {
	col := r.Table.Cols[cellIdx]
	switch col.Type {
	case "boolean":
		return cell.Value, nil
	case "number":
		return cell.Value, nil
	case "string":
		// `string` type does not have the raw value
		return cell.Value, nil
	case "date", "datetime", "timeofday":
		return cell.Raw, nil
	}
	return nil, fmt.Errorf("unsupported cell value: %s", col.Type)
}

type rawQueryRowsResultTable struct {
	Cols []rawQueryRowsResultColumn `json:"cols"`
	Rows []rawQueryRowsResultRow    `json:"rows"`
}

type rawQueryRowsResultColumn struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type rawQueryRowsResultRow struct {
	Cells []rawQueryRowsResultCell `json:"c"`
}

type rawQueryRowsResultCell struct {
	Value interface{} `json:"v"`
	Raw   string      `json:"f"`
}

type QueryRowsResult struct {
	Rows [][]interface{}
}
