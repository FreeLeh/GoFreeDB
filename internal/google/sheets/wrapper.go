package sheets

import (
	"context"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type appendMode string

const (
	majorDimensionRows                      = "ROWS"
	valueInputUserEntered                   = "USER_ENTERED"
	responseValueRenderFormatted            = "FORMATTED_VALUE"
	appendModeInsert             appendMode = "INSERT_ROWS"
	appendModeOverwrite          appendMode = "OVERWRITE"
)

type AuthClient interface {
	HTTPClient() *http.Client
}

type Wrapper struct {
	service *sheets.Service
}

func (w *Wrapper) CreateSpreadsheet(ctx context.Context, title string) (string, error) {
	createSpreadsheetReq := w.service.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: title},
	}).Context(ctx)

	spreadsheet, err := createSpreadsheetReq.Do()
	if err != nil {
		return "", err
	}
	return spreadsheet.SpreadsheetId, nil
}

func (w *Wrapper) CreateSheet(ctx context.Context, spreadsheetID string, sheetName string) error {
	addSheetReq := &sheets.AddSheetRequest{Properties: &sheets.SheetProperties{Title: sheetName}}
	requests := []*sheets.Request{
		{AddSheet: addSheetReq},
	}
	batchUpdateSpreadsheetReq := w.service.Spreadsheets.BatchUpdate(
		spreadsheetID,
		&sheets.BatchUpdateSpreadsheetRequest{Requests: requests},
	).Context(ctx)

	_, err := batchUpdateSpreadsheetReq.Do()
	return err
}

func (w *Wrapper) InsertRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (InsertRowsResult, error) {
	return w.insertRows(ctx, spreadsheetID, a1Range, values, appendModeInsert)
}

func (w *Wrapper) OverwriteRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (InsertRowsResult, error) {
	return w.insertRows(ctx, spreadsheetID, a1Range, values, appendModeOverwrite)
}

func (w *Wrapper) insertRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}, mode appendMode) (InsertRowsResult, error) {
	valueRange := &sheets.ValueRange{
		MajorDimension: majorDimensionRows,
		Range:          a1Range,
		Values:         values,
	}

	req := w.service.Spreadsheets.Values.Append(spreadsheetID, a1Range, valueRange).
		InsertDataOption(string(mode)).
		IncludeValuesInResponse(true).
		ResponseValueRenderOption(responseValueRenderFormatted).
		ValueInputOption(valueInputUserEntered).
		Context(ctx)

	resp, err := req.Do()
	if err != nil {
		return InsertRowsResult{}, err
	}

	return InsertRowsResult{
		UpdatedRange:   NewA1Range(resp.Updates.UpdatedRange),
		UpdatedRows:    resp.Updates.UpdatedRows,
		UpdatedColumns: resp.Updates.UpdatedColumns,
		UpdatedCells:   resp.Updates.UpdatedCells,
		InsertedValues: resp.Updates.UpdatedData.Values,
	}, nil
}

func (w *Wrapper) UpdateRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (UpdateRowsResult, error) {
	valueRange := &sheets.ValueRange{
		MajorDimension: majorDimensionRows,
		Range:          a1Range,
		Values:         values,
	}

	req := w.service.Spreadsheets.Values.Update(spreadsheetID, a1Range, valueRange).
		IncludeValuesInResponse(true).
		ResponseValueRenderOption(responseValueRenderFormatted).
		ValueInputOption(valueInputUserEntered).
		Context(ctx)

	resp, err := req.Do()
	if err != nil {
		return UpdateRowsResult{}, err
	}

	return UpdateRowsResult{
		UpdatedRange:   NewA1Range(resp.UpdatedRange),
		UpdatedRows:    resp.UpdatedRows,
		UpdatedColumns: resp.UpdatedColumns,
		UpdatedCells:   resp.UpdatedCells,
		UpdatedValues:  resp.UpdatedData.Values,
	}, nil
}

func (w *Wrapper) Clear(ctx context.Context, spreadsheetID string, ranges []string) ([]string, error) {
	req := w.service.Spreadsheets.Values.BatchClear(spreadsheetID, &sheets.BatchClearValuesRequest{Ranges: ranges}).
		Context(ctx)
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}
	return resp.ClearedRanges, nil
}

func NewWrapper(authClient AuthClient) (*Wrapper, error) {
	// The `ctx` provided into `NewService` is not really used for anything in our case.
	// Internally it seems it's used for creating a new HTTP client, but we already provide with our
	// own auth HTTP client.
	service, err := sheets.NewService(context.Background(), option.WithHTTPClient(authClient.HTTPClient()))
	if err != nil {
		return nil, err
	}
	return &Wrapper{service: service}, nil
}
