package store

import (
	"context"
	"github.com/FreeLeh/GoFreeDB/internal/lark/sheets"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type MockWrapper struct {
	CreateSheetError error

	GetSheetsResult sheets.GetSheetsResult
	GetSheetsError  error

	DeleteSheetsError error

	InsertRowsResult sheets.InsertRowsResult
	InsertRowsError  error

	OverwriteRowsResult sheets.InsertRowsResult
	OverwriteRowsError  error

	BatchUpdateRowsResult []sheets.BatchUpdateRowsResult
	BatchUpdateRowsError  error

	QueryRowsResult sheets.QueryRowsResult
	QueryRowsError  error

	ClearError error
}

func (w *MockWrapper) CreateSheet(
	ctx context.Context,
	spreadsheetToken string,
	sheetName string,
) error {
	return w.CreateSheetError
}

func (w *MockWrapper) GetSheets(
	ctx context.Context,
	spreadsheetToken string,
) (sheets.GetSheetsResult, error) {
	return w.GetSheetsResult, w.GetSheetsError
}

func (w *MockWrapper) DeleteSheets(
	ctx context.Context,
	spreadsheetToken string,
	sheetIDs []string,
) error {
	return w.DeleteSheetsError
}

func (w *MockWrapper) QueryRows(
	ctx context.Context,
	spreadsheetToken string,
	a1Range models.A1Range,
	query string,
) (sheets.QueryRowsResult, error) {
	return w.QueryRowsResult, w.QueryRowsError
}

func (w *MockWrapper) OverwriteRows(
	ctx context.Context,
	spreadsheetToken string,
	a1Range models.A1Range,
	values [][]interface{},
) (sheets.InsertRowsResult, error) {
	return w.OverwriteRowsResult, w.OverwriteRowsError
}

func (w *MockWrapper) BatchUpdateRows(
	ctx context.Context,
	spreadsheetToken string,
	requests []sheets.BatchUpdateRowsRequest,
) ([]sheets.BatchUpdateRowsResult, error) {
	return w.BatchUpdateRowsResult, w.BatchUpdateRowsError
}

func (w *MockWrapper) Clear(
	ctx context.Context,
	spreadsheetToken string,
	ranges []models.A1Range,
) error {
	return w.ClearError
}
