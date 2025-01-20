package sheets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"io"
	"net/http"
)

//go:generate mockgen -destination=wrapper_mock.go -package=sheets -build_constraint=gofreedb_test . AccessTokenGetter
type AccessTokenGetter interface {
	AccessToken() (string, error)
}

type Wrapper struct {
	accessTokenGetter AccessTokenGetter
	httpClient        *http.Client
}

func (w *Wrapper) CreateSheet(ctx context.Context, spreadsheetToken string, sheetName string) error {
	url := fmt.Sprintf(createSheetURL, spreadsheetToken)
	req := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"addSheet": map[string]interface{}{
					"properties": map[string]interface{}{
						"title": sheetName,
					},
				},
			},
		},
	}

	var resp baseHTTPResp[struct{}]
	if err := w.callAPI(
		ctx,
		http.MethodPost,
		"lark_create_sheet",
		url,
		req,
		&resp,
	); err != nil {
		return err
	}

	if resp.Code != apiStatusCodeOK {
		return fmt.Errorf(
			"failed calling CreateSheet, non-zero resp code, req: %s, resp: %s",
			common.JSONEncodeNoError(req),
			common.JSONEncodeNoError(resp),
		)
	}
	return nil
}

func (w *Wrapper) GetSheets(
	ctx context.Context,
	spreadsheetToken string,
) (GetSheetsResult, error) {
	url := fmt.Sprintf(getSheetsURL, spreadsheetToken)

	var resp baseHTTPResp[getSheetsHTTPResp]
	if err := w.callAPI(
		ctx,
		http.MethodGet,
		"lark_get_sheets",
		url,
		nil,
		&resp,
	); err != nil {
		return GetSheetsResult{}, err
	}

	if resp.Code != apiStatusCodeOK {
		return GetSheetsResult{}, fmt.Errorf(
			"failed calling insertRows, non-zero resp code, req: nil, resp: %s",
			common.JSONEncodeNoError(resp),
		)
	}
	return GetSheetsResult{Sheets: resp.Data.Sheets}, nil
}

func (w *Wrapper) DeleteSheets(ctx context.Context, spreadsheetToken string, sheetIDs []string) error {
	url := fmt.Sprintf(deleteSheetsURL, spreadsheetToken)

	deleteReq := make([]map[string]interface{}, 0, len(sheetIDs))
	for _, sheetID := range sheetIDs {
		deleteReq = append(deleteReq, map[string]interface{}{
			"deleteSheet": map[string]interface{}{
				"sheetId": sheetID,
			},
		})
	}

	req := map[string]interface{}{
		"requests": deleteReq,
	}

	var resp baseHTTPResp[struct{}]
	if err := w.callAPI(
		ctx,
		http.MethodPost,
		"lark_delete_sheets",
		url,
		req,
		&resp,
	); err != nil {
		return err
	}

	if resp.Code != apiStatusCodeOK {
		return fmt.Errorf(
			"failed calling DeleteSheets, non-zero resp code, req: %s, resp: %s",
			common.JSONEncodeNoError(req),
			common.JSONEncodeNoError(resp),
		)
	}
	return nil
}

func (w *Wrapper) OverwriteRows(
	ctx context.Context,
	spreadsheetToken string,
	a1Range models.A1Range,
	values [][]interface{},
) (InsertRowsResult, error) {
	return w.insertRows(ctx, spreadsheetToken, a1Range, values, appendModeOverwrite)
}

func (w *Wrapper) insertRows(
	ctx context.Context,
	spreadsheetToken string,
	a1Range models.A1Range,
	values [][]interface{},
	mode appendMode,
) (InsertRowsResult, error) {
	url := fmt.Sprintf(appendValuesURL, spreadsheetToken, mode)
	req := map[string]interface{}{
		"valueRange": map[string]interface{}{
			"range":  a1Range.Original,
			"values": values,
		},
	}

	var resp baseHTTPResp[insertRowsHTTPResp]
	if err := w.callAPI(
		ctx,
		http.MethodPost,
		fmt.Sprintf("lark_insert_rows_%s", mode),
		url,
		req,
		&resp,
	); err != nil {
		return InsertRowsResult{}, err
	}

	if resp.Code != apiStatusCodeOK {
		return InsertRowsResult{}, fmt.Errorf(
			"failed calling insertRows, non-zero resp code, req: %s, resp: %s",
			common.JSONEncodeNoError(req),
			common.JSONEncodeNoError(resp),
		)
	}
	return InsertRowsResult{
		UpdatedRange:   models.NewA1RangeFromString(resp.Data.Updates.UpdatedRange),
		UpdatedRows:    resp.Data.Updates.UpdatedRows,
		UpdatedColumns: resp.Data.Updates.UpdatedColumns,
		UpdatedCells:   resp.Data.Updates.UpdatedCells,
	}, nil
}

func (w *Wrapper) BatchUpdateRows(
	ctx context.Context,
	spreadsheetToken string,
	requests []BatchUpdateRowsRequest,
) error {
	return w.batchUpdateRows(ctx, "lark_batch_update_rows", spreadsheetToken, requests)
}

func (w *Wrapper) Clear(
	ctx context.Context,
	spreadsheetToken string,
	ranges []models.A1Range,
) error {
	requests := make([]BatchUpdateRowsRequest, 0, len(ranges))
	for _, rng := range ranges {
		values := make([][]interface{}, 0, rng.NumRows())

		for row := 0; row < rng.NumRows(); row++ {
			rowValues := make([]interface{}, 0, rng.NumCols())
			for col := 0; col < rng.NumCols(); col++ {
				rowValues = append(rowValues, "")
			}
			values = append(values, rowValues)
		}

		requests = append(requests, BatchUpdateRowsRequest{
			A1Range: rng,
			Values:  values,
		})
	}
	return w.batchUpdateRows(ctx, "lark_clear", spreadsheetToken, requests)
}

func (w *Wrapper) batchUpdateRows(
	ctx context.Context,
	name string,
	spreadsheetToken string,
	requests []BatchUpdateRowsRequest,
) error {
	url := fmt.Sprintf(batchUpdateRowsURL, spreadsheetToken)

	req := make([]map[string]interface{}, 0, len(requests))
	for _, r := range requests {
		valuesPerReq := make([][]interface{}, 0, len(r.Values))

		for _, v := range r.Values {
			valuesPerRow := make([]interface{}, 0, len(v))
			for _, item := range v {
				valuesPerRow = append(valuesPerRow, item)
			}
			valuesPerReq = append(valuesPerReq, valuesPerRow)
		}

		req = append(req, map[string]interface{}{
			"range":  r.A1Range.Original,
			"values": valuesPerReq,
		})
	}

	var resp baseHTTPResp[struct{}]
	if err := w.callAPI(
		ctx,
		http.MethodPost,
		name,
		url,
		req,
		&resp,
	); err != nil {
		return err
	}

	if resp.Code != apiStatusCodeOK {
		return fmt.Errorf(
			"failed calling batchUpdateRows (%s), non-zero resp code, req: %s, resp: %s",
			name,
			common.JSONEncodeNoError(req),
			common.JSONEncodeNoError(resp),
		)
	}
	return nil
}

func (w *Wrapper) callAPI(
	ctx context.Context,
	method string,
	apiName string,
	url string,
	req any,
	resp any,
) error {
	var reqJSON []byte
	if req != nil {
		raw, err := json.Marshal(req)
		if err != nil {
			return err
		}
		reqJSON = raw
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf(
			"failed to create request, name: %s, req: %s, err: %s",
			apiName,
			string(reqJSON),
			err,
		)
	}

	token, err := w.accessTokenGetter.AccessToken()
	if err != nil {
		return fmt.Errorf(
			"call API failed when getting access token, name: %s, req: %s, err: %s",
			apiName,
			string(reqJSON),
			err,
		)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf(
			"call API failed when doing http request, name: %s, req: %s, err: %s",
			apiName,
			string(reqJSON),
			err,
		)
	}
	defer httpResp.Body.Close()

	respJSON, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf(
			"call API failed when reading HTTP response body, name: %s, req: %s, resp: %s, err: %s",
			apiName,
			string(reqJSON),
			string(respJSON),
			err,
		)
	}
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"call API http status not OK, name: %s, req: %s, resp: %s",
			apiName,
			string(reqJSON),
			string(respJSON),
		)
	}

	if err = json.Unmarshal(respJSON, resp); err != nil {
		return fmt.Errorf(
			"call API failed when unmarshal response body, name: %s, req: %s, resp: %s, err: %s",
			apiName,
			string(reqJSON),
			string(respJSON),
			err,
		)
	}
	return nil
}

func NewWrapper(accessTokenGetter AccessTokenGetter) *Wrapper {
	return &Wrapper{
		accessTokenGetter: accessTokenGetter,
		httpClient:        http.DefaultClient,
	}
}
