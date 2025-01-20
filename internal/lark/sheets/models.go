package sheets

import "github.com/FreeLeh/GoFreeDB/internal/models"

type appendMode string

const (
	appendModeInsert    appendMode = "INSERT_ROWS"
	appendModeOverwrite appendMode = "OVERWRITE"

	createSheetURL     = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/sheets_batch_update"
	deleteSheetsURL    = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/sheets_batch_update"
	appendValuesURL    = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/values_append?insertDataOption=%s"
	getSheetsURL       = "https://open.larksuite.com/open-apis/sheets/v3/spreadsheets/%s/sheets/query"
	batchUpdateRowsURL = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/values_batch_update"

	apiStatusCodeOK = 0
)

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
