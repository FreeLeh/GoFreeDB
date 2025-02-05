package google

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

// SheetSelectStmt encapsulates information required to query the row store.
type SheetSelectStmt struct {
	store        *SheetRowStore
	columns      []string
	queryBuilder *common.QueryBuilder
	output       interface{}
}

// Where specifies the condition to meet for a row to be included.
//
// "condition" specifies the WHERE clause.
// Values in the WHERE clause should be replaced by a placeholder "?".
// The actual values used for each placeholder (ordering matters) are provided via the "args" parameter.
//
// "args" specifies the real value to replace each placeholder in the WHERE clause.
// Note that the first "args" value will replace the first placeholder "?" in the WHERE clause.
//
// If you want to understand the reason behind this design, please read the protocol page: https://github.com/FreeLeh/docs/blob/main/freedb/protocols.md.
//
// All conditions supported by Google Sheet "QUERY" function are supported by this library.
// You can read the full information in https://developers.google.com/chart/interactive/docs/querylanguage#where.
func (s *SheetSelectStmt) Where(condition string, args ...interface{}) *SheetSelectStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// OrderBy specifies the column ordering.
//
// The default value is no ordering specified.
func (s *SheetSelectStmt) OrderBy(ordering []models.ColumnOrderBy) *SheetSelectStmt {
	s.queryBuilder.OrderBy(ordering)
	return s
}

// Limit specifies the number of rows to retrieve.
//
// The default value is 0.
func (s *SheetSelectStmt) Limit(limit uint64) *SheetSelectStmt {
	s.queryBuilder.Limit(limit)
	return s
}

// Offset specifies the number of rows to skip before starting to include the rows.
//
// The default value is 0.
func (s *SheetSelectStmt) Offset(offset uint64) *SheetSelectStmt {
	s.queryBuilder.Offset(offset)
	return s
}

// Exec retrieves rows matching with the given condition.
//
// There is only 1 API call behind the scene.
func (s *SheetSelectStmt) Exec(ctx context.Context) error {
	if err := s.ensureOutputSlice(); err != nil {
		return err
	}

	stmt, err := s.queryBuilder.Generate()
	if err != nil {
		return err
	}

	result, err := s.store.wrapper.QueryRows(
		ctx,
		s.store.spreadsheetID,
		s.store.sheetName,
		stmt,
		true,
	)
	if err != nil {
		return err
	}

	m := s.buildQueryResultMap(result)
	return common.MapStructureDecode(m, s.output)
}

func (s *SheetSelectStmt) buildQueryResultMap(original QueryRowsResult) []map[string]interface{} {
	result := make([]map[string]interface{}, len(original.Rows))

	for rowIdx, row := range original.Rows {
		result[rowIdx] = make(map[string]interface{}, len(row))

		for colIdx, value := range row {
			col := s.columns[colIdx]
			result[rowIdx][col] = value
		}
	}

	return result
}

func (s *SheetSelectStmt) ensureOutputSlice() error {
	// Passing an uninitialised slice will not compare to nil due to this: https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/
	// Only if passing an untyped `nil` will compare to the `nil` in the line below.
	// Observations as below:
	//
	// var o []int
	// o == nil --> this is true because the compiler knows `o` is nil and of type `[]int`, so the `nil` on the right side is of the same `[]int` type.
	//
	// var x interface{} = o
	// x == nil --> this is false because `o` has been boxed by `x` and the `nil` on the right side is of type `nil` (i.e. nil value of nil type).
	// x == []int(nil) --> this is true because the `nil` has been casted explicitly to `nil` of type `[]int`.
	if s.output == nil {
		return errors.New("select statement output cannot be empty or nil")
	}

	t := reflect.TypeOf(s.output)
	if t.Kind() != reflect.Ptr {
		return errors.New("select statement output must be a pointer to a slice of something")
	}

	elem := t.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("select statement output must be a pointer to a slice of something; current output type: %s", t.Kind().String())
	}

	return nil
}

func newSheetSelectStmt(store *SheetRowStore, output interface{}, columns []string) *SheetSelectStmt {
	if len(columns) == 0 {
		columns = store.config.Columns
	}

	return &SheetSelectStmt{
		store:   store,
		columns: columns,
		queryBuilder: common.NewQueryBuilder(
			store.colsMapping.NameMap(),
			ridWhereClauseInterceptor,
			columns,
		),
		output: output,
	}
}

// SheetInsertStmt encapsulates information required to insert new rows into the Google Sheet.
type SheetInsertStmt struct {
	store *SheetRowStore
	rows  []interface{}
}

func (s *SheetInsertStmt) convertRowToSlice(row interface{}) ([]interface{}, error) {
	if row == nil {
		return nil, errors.New("row type must not be nil")
	}

	t := reflect.TypeOf(row)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("row type must be either a struct or a slice")
	}

	var output map[string]interface{}
	if err := common.MapStructureDecode(row, &output); err != nil {
		return nil, err
	}

	result := make([]interface{}, len(s.store.colsMapping))
	result[0] = rowIdxFormula

	for col, value := range output {
		if colIdx, ok := s.store.colsMapping[col]; ok {
			escapedValue, err := common.EscapeValue(col, value, s.store.colsWithFormula)
			if err != nil {
				return nil, err
			}
			if err = common.CheckIEEE754SafeInteger(escapedValue); err != nil {
				return nil, err
			}
			result[colIdx.Idx] = escapedValue
		}
	}

	return result, nil
}

// Exec inserts the provided new rows data into Google Sheet.
// This method calls the relevant Google Sheet APIs to actually insert the new rows.
//
// There is only 1 API call behind the scene.
func (s *SheetInsertStmt) Exec(ctx context.Context) error {
	if len(s.rows) == 0 {
		return nil
	}

	convertedRows := make([][]interface{}, 0, len(s.rows))
	for _, row := range s.rows {
		r, err := s.convertRowToSlice(row)
		if err != nil {
			return fmt.Errorf("cannot execute google sheet insert statement due to row conversion error: %w", err)
		}
		convertedRows = append(convertedRows, r)
	}

	_, err := s.store.wrapper.OverwriteRows(
		ctx,
		s.store.spreadsheetID,
		models.NewA1Range(s.store.sheetName, defaultRowFullTableRange),
		convertedRows,
	)
	return err
}

func newGoogleSheetInsertStmt(store *SheetRowStore, rows []interface{}) *SheetInsertStmt {
	return &SheetInsertStmt{
		store: store,
		rows:  rows,
	}
}

// SheetUpdateStmt encapsulates information required to update rows.
type SheetUpdateStmt struct {
	store        *SheetRowStore
	colToValue   map[string]interface{}
	queryBuilder *common.QueryBuilder
}

// Where specifies the condition to choose which rows are affected.
//
// It works just like the SheetSelectStmt.Where() method.
// Please read SheetSelectStmt.Where() for more details.
func (s *SheetUpdateStmt) Where(condition string, args ...interface{}) *SheetUpdateStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// Exec updates rows matching the condition with the new values for affected columns.
//
// There are 2 API calls behind the scene.
func (s *SheetUpdateStmt) Exec(ctx context.Context) error {
	if len(s.colToValue) == 0 {
		return errors.New("empty colToValue, at least one column must be updated")
	}

	selectStmt, err := s.queryBuilder.Generate()
	if err != nil {
		return err
	}

	indices, err := getRowIndices(ctx, s.store, selectStmt)
	if err != nil {
		return err
	}
	if len(indices) == 0 {
		return nil
	}

	requests, err := s.generateBatchUpdateRequests(indices)
	if err != nil {
		return err
	}

	_, err = s.store.wrapper.BatchUpdateRows(ctx, s.store.spreadsheetID, requests)
	return err
}

func (s *SheetUpdateStmt) generateBatchUpdateRequests(rowIndices []int64) ([]BatchUpdateRowsRequest, error) {
	requests := make([]BatchUpdateRowsRequest, 0)

	for col, value := range s.colToValue {
		colIdx, ok := s.store.colsMapping[col]
		if !ok {
			return nil, fmt.Errorf("failed to update, unknown column name provided: %s", col)
		}

		escapedValue, err := common.EscapeValue(col, value, s.store.colsWithFormula)
		if err != nil {
			return nil, err
		}
		if err = common.CheckIEEE754SafeInteger(escapedValue); err != nil {
			return nil, err
		}

		for _, rowIdx := range rowIndices {
			a1Range := colIdx.Name + strconv.FormatInt(rowIdx, 10)
			requests = append(requests, BatchUpdateRowsRequest{
				A1Range: models.NewA1Range(s.store.sheetName, a1Range),
				Values:  [][]interface{}{{escapedValue}},
			})
		}
	}

	return requests, nil
}

func newSheetUpdateStmt(store *SheetRowStore, colToValue map[string]interface{}) *SheetUpdateStmt {
	return &SheetUpdateStmt{
		store:      store,
		colToValue: colToValue,
		queryBuilder: common.NewQueryBuilder(
			store.colsMapping.NameMap(),
			ridWhereClauseInterceptor,
			[]string{rowIdxCol},
		),
	}
}

// SheetDeleteStmt encapsulates information required to delete rows.
type SheetDeleteStmt struct {
	store        *SheetRowStore
	queryBuilder *common.QueryBuilder
}

// Where specifies the condition to choose which rows are affected.
//
// It works just like the SheetSelectStmt.Where() method.
// Please read SheetSelectStmt.Where() for more details.
func (s *SheetDeleteStmt) Where(condition string, args ...interface{}) *SheetDeleteStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// Exec deletes rows matching the condition.
//
// There are 2 API calls behind the scene.
func (s *SheetDeleteStmt) Exec(ctx context.Context) error {
	selectStmt, err := s.queryBuilder.Generate()
	if err != nil {
		return err
	}

	indices, err := getRowIndices(ctx, s.store, selectStmt)
	if err != nil {
		return err
	}
	if len(indices) == 0 {
		return nil
	}

	_, err = s.store.wrapper.Clear(ctx, s.store.spreadsheetID, generateRowA1Ranges(s.store.sheetName, indices))
	return err
}

func newSheetDeleteStmt(store *SheetRowStore) *SheetDeleteStmt {
	return &SheetDeleteStmt{
		store: store,
		queryBuilder: common.NewQueryBuilder(
			store.colsMapping.NameMap(),
			ridWhereClauseInterceptor,
			[]string{rowIdxCol},
		),
	}
}

// SheetCountStmt encapsulates information required to count the number of rows matching some conditions.
type SheetCountStmt struct {
	store        *SheetRowStore
	queryBuilder *common.QueryBuilder
}

// Where specifies the condition to choose which rows are affected.
//
// It works just like the SheetSelectStmt.Where() method.
// Please read SheetSelectStmt.Where() for more details.
func (s *SheetCountStmt) Where(condition string, args ...interface{}) *SheetCountStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// Exec counts the number of rows matching the provided condition.
//
// There is only 1 API call behind the scene.
func (s *SheetCountStmt) Exec(ctx context.Context) (uint64, error) {
	selectStmt, err := s.queryBuilder.Generate()
	if err != nil {
		return 0, err
	}

	result, err := s.store.wrapper.QueryRows(ctx, s.store.spreadsheetID, s.store.sheetName, selectStmt, true)
	if err != nil {
		return 0, err
	}

	if len(result.Rows) != 1 || len(result.Rows[0]) != 1 {
		return 0, errors.New("")
	}

	count := result.Rows[0][0].(float64)
	return uint64(count), nil
}

func newSheetCountStmt(store *SheetRowStore) *SheetCountStmt {
	countClause := fmt.Sprintf("COUNT(%s)", rowIdxCol)
	return &SheetCountStmt{
		store: store,
		queryBuilder: common.NewQueryBuilder(
			store.colsMapping.NameMap(),
			ridWhereClauseInterceptor,
			[]string{countClause},
		),
	}
}

func getRowIndices(ctx context.Context, store *SheetRowStore, selectStmt string) ([]int64, error) {
	result, err := store.wrapper.QueryRows(ctx, store.spreadsheetID, store.sheetName, selectStmt, true)
	if err != nil {
		return nil, err
	}
	if len(result.Rows) == 0 {
		return nil, nil
	}

	rowIndices := make([]int64, 0)
	for _, row := range result.Rows {
		if len(row) != 1 {
			return nil, fmt.Errorf("error retrieving row indices: %+v", result)
		}

		idx, ok := row[0].(float64)
		if !ok {
			return nil, fmt.Errorf("error converting row indices, value: %+v", row[0])
		}

		rowIndices = append(rowIndices, int64(idx))
	}

	return rowIndices, nil
}

func generateRowA1Ranges(sheetName string, indices []int64) []models.A1Range {
	locations := make([]models.A1Range, len(indices))
	for i := range indices {
		locations[i] = models.NewA1Range(
			sheetName,
			fmt.Sprintf(rowDeleteRangeTemplate, indices[i], indices[i]),
		)
	}
	return locations
}

func ridWhereClauseInterceptor(where string) string {
	if where == "" {
		return rowWhereEmptyConditionTemplate
	}
	return fmt.Sprintf(rowWhereNonEmptyConditionTemplate, where)
}
