package sheets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type AuthClient interface {
	HTTPClient() *http.Client
}

type Wrapper struct {
	service   *sheets.Service
	rawClient *http.Client
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

func (w *Wrapper) GetSheetNameToID(ctx context.Context, spreadsheetID string) (map[string]int64, error) {
	resp, err := w.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, sheet := range resp.Sheets {
		if sheet.Properties == nil {
			return nil, errors.New("failed getSheetIDByName due to empty sheet properties")
		}
		result[sheet.Properties.Title] = sheet.Properties.SheetId
	}

	return result, nil
}

func (w *Wrapper) DeleteSheets(ctx context.Context, spreadsheetID string, sheetIDs []int64) error {
	requests := make([]*sheets.Request, 0, len(sheetIDs))
	for _, sheetID := range sheetIDs {
		deleteSheetReq := &sheets.DeleteSheetRequest{SheetId: sheetID}
		requests = append(requests, &sheets.Request{DeleteSheet: deleteSheetReq})
	}

	batchUpdateSpreadsheetReq := w.service.Spreadsheets.BatchUpdate(
		spreadsheetID,
		&sheets.BatchUpdateSpreadsheetRequest{Requests: requests},
	).Context(ctx)

	_, err := batchUpdateSpreadsheetReq.Do()
	return err
}

func (w *Wrapper) InsertRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range string,
	values [][]interface{},
) (InsertRowsResult, error) {
	return w.insertRows(ctx, spreadsheetID, a1Range, values, appendModeInsert)
}

func (w *Wrapper) OverwriteRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range string,
	values [][]interface{},
) (InsertRowsResult, error) {
	return w.insertRows(ctx, spreadsheetID, a1Range, values, appendModeOverwrite)
}

func (w *Wrapper) insertRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range string,
	values [][]interface{},
	mode appendMode,
) (InsertRowsResult, error) {
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

func (w *Wrapper) UpdateRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range string,
	values [][]interface{},
) (UpdateRowsResult, error) {
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

func (w *Wrapper) BatchUpdateRows(
	ctx context.Context,
	spreadsheetID string,
	requests []BatchUpdateRowsRequest,
) (BatchUpdateRowsResult, error) {
	valueRanges := make([]*sheets.ValueRange, len(requests))
	for i := range requests {
		valueRanges[i] = &sheets.ValueRange{
			MajorDimension: majorDimensionRows,
			Range:          requests[i].A1Range,
			Values:         requests[i].Values,
		}
	}

	batchUpdate := &sheets.BatchUpdateValuesRequest{
		Data:                      valueRanges,
		IncludeValuesInResponse:   true,
		ResponseValueRenderOption: responseValueRenderFormatted,
		ValueInputOption:          valueInputUserEntered,
	}

	req := w.service.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdate).Context(ctx)

	resp, err := req.Do()
	if err != nil {
		return BatchUpdateRowsResult{}, err
	}

	results := make(BatchUpdateRowsResult, len(requests))
	for i := range resp.Responses {
		results[i] = UpdateRowsResult{
			UpdatedRange:   NewA1Range(resp.Responses[i].UpdatedRange),
			UpdatedRows:    resp.Responses[i].UpdatedRows,
			UpdatedColumns: resp.Responses[i].UpdatedColumns,
			UpdatedCells:   resp.Responses[i].UpdatedCells,
			UpdatedValues:  resp.Responses[i].UpdatedData.Values,
		}
	}

	return results, nil
}

func (w *Wrapper) QueryRows(
	ctx context.Context,
	spreadsheetID string,
	sheetName string,
	query string,
	skipHeader bool,
) (QueryRowsResult, error) {
	rawResult, err := w.execQueryRows(ctx, spreadsheetID, sheetName, query, skipHeader)
	if err != nil {
		return QueryRowsResult{}, err
	}
	return rawResult.toQueryRowsResult()
}

func (w *Wrapper) execQueryRows(
	ctx context.Context,
	spreadsheetID string,
	sheetName string,
	query string,
	skipHeader bool,
) (rawQueryRowsResult, error) {
	params := url.Values{}
	params.Add("sheet", sheetName)
	params.Add("tqx", "responseHandler:freeleh")
	params.Add("tq", query)

	header := 0
	if skipHeader {
		header = 1
	}
	params.Add("headers", strconv.FormatInt(int64(header), 10))

	url := fmt.Sprintf(queryRowsURLTemplate, spreadsheetID) + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return rawQueryRowsResult{}, err
	}

	resp, err := w.rawClient.Do(req)
	if err != nil {
		return rawQueryRowsResult{}, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rawQueryRowsResult{}, err
	}
	respString := string(respBytes)

	firstCurly := strings.Index(respString, "{")
	if firstCurly == -1 {
		return rawQueryRowsResult{}, fmt.Errorf("opening curly bracket not found: %s", respString)
	}

	lastCurly := strings.LastIndex(respString, "}")
	if lastCurly == -1 {
		return rawQueryRowsResult{}, fmt.Errorf("closing curly bracket not found: %s", respString)
	}

	result := rawQueryRowsResult{}
	if err := json.Unmarshal([]byte(respString[firstCurly:lastCurly+1]), &result); err != nil {
		return rawQueryRowsResult{}, err
	}
	return result, nil
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
	return &Wrapper{
		service:   service,
		rawClient: authClient.HTTPClient(),
	}, nil
}
