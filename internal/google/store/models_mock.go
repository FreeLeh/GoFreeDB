package store

import (
	"context"
	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type MockWrapper struct {
	CreateSpreadsheetResult string
	CreateSpreadsheetError  error

	CreateSheetError error

	InsertRowsResult sheets.InsertRowsResult
	InsertRowsError  error

	OverwriteRowsResult sheets.InsertRowsResult
	OverwriteRowsError  error

	UpdateRowsResult sheets.UpdateRowsResult
	UpdateRowsError  error

	BatchUpdateRowsResult sheets.BatchUpdateRowsResult
	BatchUpdateRowsError  error

	QueryRowsResult sheets.QueryRowsResult
	QueryRowsError  error

	ClearResult []string
	ClearError  error
}

func (w *MockWrapper) CreateSpreadsheet(
	ctx context.Context,
	title string,
) (string, error) {
	return w.CreateSpreadsheetResult, w.CreateSpreadsheetError
}

func (w *MockWrapper) GetSheetNameToID(
	ctx context.Context,
	spreadsheetID string,
) (map[string]int64, error) {
	return nil, nil
}

func (w *MockWrapper) DeleteSheets(
	ctx context.Context,
	spreadsheetID string,
	sheetIDs []int64,
) error {
	return nil
}

func (w *MockWrapper) CreateSheet(
	ctx context.Context,
	spreadsheetID string,
	sheetName string,
) error {
	return w.CreateSheetError
}

func (w *MockWrapper) InsertRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range models.A1Range,
	values [][]interface{},
) (sheets.InsertRowsResult, error) {
	return w.InsertRowsResult, w.InsertRowsError
}

func (w *MockWrapper) OverwriteRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range models.A1Range,
	values [][]interface{},
) (sheets.InsertRowsResult, error) {
	return w.OverwriteRowsResult, w.OverwriteRowsError
}

func (w *MockWrapper) UpdateRows(
	ctx context.Context,
	spreadsheetID string,
	a1Range models.A1Range,
	values [][]interface{},
) (sheets.UpdateRowsResult, error) {
	return w.UpdateRowsResult, w.UpdateRowsError
}

func (w *MockWrapper) BatchUpdateRows(
	ctx context.Context,
	spreadsheetID string,
	requests []sheets.BatchUpdateRowsRequest,
) (sheets.BatchUpdateRowsResult, error) {
	return w.BatchUpdateRowsResult, w.BatchUpdateRowsError
}

func (w *MockWrapper) QueryRows(
	ctx context.Context,
	spreadsheetID string,
	sheetName string,
	query string,
	skipHeader bool,
) (sheets.QueryRowsResult, error) {
	return w.QueryRowsResult, w.QueryRowsError
}

func (w *MockWrapper) Clear(
	ctx context.Context,
	spreadsheetID string,
	ranges []models.A1Range,
) ([]string, error) {
	return w.ClearResult, w.ClearError
}
