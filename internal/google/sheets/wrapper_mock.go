package sheets

import "context"

type MockWrapper struct {
	CreateSpreadsheetResult string
	CreateSpreadsheetError  error

	CreateSheetError error

	InsertRowsResult InsertRowsResult
	InsertRowsError  error

	OverwriteRowsResult InsertRowsResult
	OverwriteRowsError  error

	UpdateRowsResult UpdateRowsResult
	UpdateRowsError  error

	BatchUpdateRowsResult BatchUpdateRowsResult
	BatchUpdateRowsError  error

	QueryRowsResult QueryRowsResult
	QueryRowsError  error

	ClearResult []string
	ClearError  error
}

func (w *MockWrapper) CreateSpreadsheet(ctx context.Context, title string) (string, error) {
	return w.CreateSpreadsheetResult, w.CreateSpreadsheetError
}

func (w *MockWrapper) CreateSheet(ctx context.Context, spreadsheetID string, sheetName string) error {
	return w.CreateSheetError
}

func (w *MockWrapper) InsertRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (InsertRowsResult, error) {
	return w.InsertRowsResult, w.InsertRowsError
}

func (w *MockWrapper) OverwriteRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (InsertRowsResult, error) {
	return w.OverwriteRowsResult, w.OverwriteRowsError
}

func (w *MockWrapper) UpdateRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (UpdateRowsResult, error) {
	return w.UpdateRowsResult, w.UpdateRowsError
}

func (w *MockWrapper) BatchUpdateRows(ctx context.Context, spreadsheetID string, requests []BatchUpdateRowsRequest) (BatchUpdateRowsResult, error) {
	return w.BatchUpdateRowsResult, w.BatchUpdateRowsError
}

func (w *MockWrapper) QueryRows(ctx context.Context, spreadsheetID string, sheetName string, query string, skipHeader bool) (QueryRowsResult, error) {
	return w.QueryRowsResult, w.QueryRowsError
}

func (w *MockWrapper) Clear(ctx context.Context, spreadsheetID string, ranges []string) ([]string, error) {
	return w.ClearResult, w.ClearError
}
