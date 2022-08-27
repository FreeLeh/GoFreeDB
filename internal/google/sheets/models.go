package sheets

import (
	"fmt"
	"strings"
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

type A1Range struct {
	Original  string
	SheetName string
	FromCell  string
	ToCell    string
}

func NewA1Range(s string) A1Range {
	exclamationIdx := strings.Index(s, "!")
	colonIdx := strings.Index(s, ":")

	if exclamationIdx == -1 {
		if colonIdx == -1 {
			return A1Range{
				Original:  s,
				SheetName: "",
				FromCell:  s,
				ToCell:    s,
			}
		} else {
			return A1Range{
				Original:  s,
				SheetName: "",
				FromCell:  s[:colonIdx],
				ToCell:    s[colonIdx+1:],
			}
		}
	} else {
		if colonIdx == -1 {
			return A1Range{
				Original:  s,
				SheetName: s[:exclamationIdx],
				FromCell:  s[exclamationIdx+1:],
				ToCell:    s[exclamationIdx+1:],
			}
		} else {
			return A1Range{
				Original:  s,
				SheetName: s[:exclamationIdx],
				FromCell:  s[exclamationIdx+1 : colonIdx],
				ToCell:    s[colonIdx+1:],
			}
		}
	}
}

type InsertRowsResult struct {
	UpdatedRange   A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
	InsertedValues [][]interface{}
}

type UpdateRowsResult struct {
	UpdatedRange   A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
	UpdatedValues  [][]interface{}
}

type BatchUpdateRowsRequest struct {
	A1Range string
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
		return strings.ToLower(cell.Raw) == "true", nil
	case "number":
		if strings.Contains(cell.Raw, ".") {
			return cell.Value, nil
		} else {
			val := cell.Value.(float64)
			return int64(val), nil
		}
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
