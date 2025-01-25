package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"reflect"
	"strconv"
	"strings"

	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
)

type whereInterceptorFunc func(where string) string

type queryBuilder struct {
	replacer         *strings.Replacer
	columns          []string
	where            string
	whereArgs        []interface{}
	whereInterceptor whereInterceptorFunc
	orderBy          []string
	limit            uint64
	offset           uint64
}

func (q *queryBuilder) Where(condition string, args ...interface{}) *queryBuilder {
	q.where = condition
	q.whereArgs = args
	return q
}

func (q *queryBuilder) OrderBy(ordering []models.ColumnOrderBy) *queryBuilder {
	orderBy := make([]string, 0, len(ordering))
	for _, o := range ordering {
		orderBy = append(orderBy, o.Column+" "+string(o.OrderBy))
	}

	q.orderBy = orderBy
	return q
}

func (q *queryBuilder) Limit(limit uint64) *queryBuilder {
	q.limit = limit
	return q
}

func (q *queryBuilder) Offset(offset uint64) *queryBuilder {
	q.offset = offset
	return q
}

func (q *queryBuilder) Generate() (string, error) {
	stmt := &strings.Builder{}
	stmt.WriteString("select")

	if err := q.writeCols(stmt); err != nil {
		return "", err
	}
	if err := q.writeWhere(stmt); err != nil {
		return "", err
	}
	if err := q.writeOrderBy(stmt); err != nil {
		return "", err
	}
	if err := q.writeOffset(stmt); err != nil {
		return "", err
	}
	if err := q.writeLimit(stmt); err != nil {
		return "", err
	}

	return stmt.String(), nil
}

func (q *queryBuilder) writeCols(stmt *strings.Builder) error {
	stmt.WriteString(" ")

	translated := make([]string, 0, len(q.columns))
	for _, col := range q.columns {
		translated = append(translated, q.replacer.Replace(col))
	}

	stmt.WriteString(strings.Join(translated, ", "))
	return nil
}

func (q *queryBuilder) writeWhere(stmt *strings.Builder) error {
	where := q.where
	if q.whereInterceptor != nil {
		where = q.whereInterceptor(q.where)
	}

	nArgs := strings.Count(where, "?")
	if nArgs != len(q.whereArgs) {
		return fmt.Errorf("number of arguments required in the 'where' clause (%d) is not the same as the number of provided arguments (%d)", nArgs, len(q.whereArgs))
	}

	where = q.replacer.Replace(where)
	tokens := strings.Split(where, "?")

	result := make([]string, 0)
	result = append(result, strings.TrimSpace(tokens[0]))

	for i, token := range tokens[1:] {
		arg, err := q.convertArg(q.whereArgs[i])
		if err != nil {
			return fmt.Errorf("failed converting 'where' arguments: %v, %w", arg, err)
		}
		result = append(result, arg, strings.TrimSpace(token))
	}

	stmt.WriteString(" where ")
	stmt.WriteString(strings.Join(result, " "))
	return nil
}

func (q *queryBuilder) convertArg(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return q.convertInt(arg)
	case float32, float64:
		return q.convertFloat(arg)
	case string, []byte:
		return q.convertString(arg)
	case bool:
		return strconv.FormatBool(converted), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *queryBuilder) convertInt(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case int:
		return strconv.FormatInt(int64(converted), 10), nil
	case int8:
		return strconv.FormatInt(int64(converted), 10), nil
	case int16:
		return strconv.FormatInt(int64(converted), 10), nil
	case int32:
		return strconv.FormatInt(int64(converted), 10), nil
	case int64:
		return strconv.FormatInt(converted, 10), nil
	case uint:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint64:
		return strconv.FormatUint(converted, 10), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *queryBuilder) convertFloat(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case float32:
		return strconv.FormatFloat(float64(converted), 'f', -1, 64), nil
	case float64:
		return strconv.FormatFloat(converted, 'f', -1, 64), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *queryBuilder) convertString(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case string:
		cleaned := strings.ToLower(strings.TrimSpace(converted))
		if googleSheetSelectStmtStringKeyword.MatchString(cleaned) {
			return converted, nil
		}
		return strconv.Quote(converted), nil
	case []byte:
		return strconv.Quote(string(converted)), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *queryBuilder) writeOrderBy(stmt *strings.Builder) error {
	if len(q.orderBy) == 0 {
		return nil
	}

	stmt.WriteString(" order by ")
	result := make([]string, 0, len(q.orderBy))

	for _, o := range q.orderBy {
		result = append(result, q.replacer.Replace(o))
	}

	stmt.WriteString(strings.Join(result, ", "))
	return nil
}

func (q *queryBuilder) writeOffset(stmt *strings.Builder) error {
	if q.offset == 0 {
		return nil
	}

	stmt.WriteString(" offset ")
	stmt.WriteString(strconv.FormatInt(int64(q.offset), 10))
	return nil
}

func (q *queryBuilder) writeLimit(stmt *strings.Builder) error {
	if q.limit == 0 {
		return nil
	}

	stmt.WriteString(" limit ")
	stmt.WriteString(strconv.FormatInt(int64(q.limit), 10))
	return nil
}

func newQueryBuilder(
	colReplacements map[string]string,
	whereInterceptor whereInterceptorFunc,
	colSelected []string,
) *queryBuilder {
	replacements := make([]string, 0, 2*len(colReplacements))
	for col, repl := range colReplacements {
		replacements = append(replacements, col, repl)
	}

	return &queryBuilder{
		replacer:         strings.NewReplacer(replacements...),
		columns:          colSelected,
		whereInterceptor: whereInterceptor,
	}
}

// GoogleSheetSelectStmt encapsulates information required to query the row store.
type GoogleSheetSelectStmt struct {
	store        *GoogleSheetRowStore
	columns      []string
	queryBuilder *queryBuilder
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
func (s *GoogleSheetSelectStmt) Where(condition string, args ...interface{}) *GoogleSheetSelectStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// OrderBy specifies the column ordering.
//
// The default value is no ordering specified.
func (s *GoogleSheetSelectStmt) OrderBy(ordering []models.ColumnOrderBy) *GoogleSheetSelectStmt {
	s.queryBuilder.OrderBy(ordering)
	return s
}

// Limit specifies the number of rows to retrieve.
//
// The default value is 0.
func (s *GoogleSheetSelectStmt) Limit(limit uint64) *GoogleSheetSelectStmt {
	s.queryBuilder.Limit(limit)
	return s
}

// Offset specifies the number of rows to skip before starting to include the rows.
//
// The default value is 0.
func (s *GoogleSheetSelectStmt) Offset(offset uint64) *GoogleSheetSelectStmt {
	s.queryBuilder.Offset(offset)
	return s
}

// Exec retrieves rows matching with the given condition.
//
// There is only 1 API call behind the scene.
func (s *GoogleSheetSelectStmt) Exec(ctx context.Context) error {
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

func (s *GoogleSheetSelectStmt) buildQueryResultMap(original sheets.QueryRowsResult) []map[string]interface{} {
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

func (s *GoogleSheetSelectStmt) ensureOutputSlice() error {
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

func newGoogleSheetSelectStmt(store *GoogleSheetRowStore, output interface{}, columns []string) *GoogleSheetSelectStmt {
	if len(columns) == 0 {
		columns = store.config.Columns
	}

	return &GoogleSheetSelectStmt{
		store:        store,
		columns:      columns,
		queryBuilder: newQueryBuilder(store.colsMapping.NameMap(), ridWhereClauseInterceptor, columns),
		output:       output,
	}
}

// GoogleSheetInsertStmt encapsulates information required to insert new rows into the Google Sheet.
type GoogleSheetInsertStmt struct {
	store *GoogleSheetRowStore
	rows  []interface{}
}

func (s *GoogleSheetInsertStmt) convertRowToSlice(row interface{}) ([]interface{}, error) {
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
			escapedValue, err := escapeValue(col, value, s.store.colsWithFormula)
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
func (s *GoogleSheetInsertStmt) Exec(ctx context.Context) error {
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
		common.GetA1Range(s.store.sheetName, defaultRowFullTableRange),
		convertedRows,
	)
	return err
}

func newGoogleSheetInsertStmt(store *GoogleSheetRowStore, rows []interface{}) *GoogleSheetInsertStmt {
	return &GoogleSheetInsertStmt{
		store: store,
		rows:  rows,
	}
}

// GoogleSheetUpdateStmt encapsulates information required to update rows.
type GoogleSheetUpdateStmt struct {
	store        *GoogleSheetRowStore
	colToValue   map[string]interface{}
	queryBuilder *queryBuilder
}

// Where specifies the condition to choose which rows are affected.
//
// It works just like the GoogleSheetSelectStmt.Where() method.
// Please read GoogleSheetSelectStmt.Where() for more details.
func (s *GoogleSheetUpdateStmt) Where(condition string, args ...interface{}) *GoogleSheetUpdateStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// Exec updates rows matching the condition with the new values for affected columns.
//
// There are 2 API calls behind the scene.
func (s *GoogleSheetUpdateStmt) Exec(ctx context.Context) error {
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

func (s *GoogleSheetUpdateStmt) generateBatchUpdateRequests(rowIndices []int64) ([]sheets.BatchUpdateRowsRequest, error) {
	requests := make([]sheets.BatchUpdateRowsRequest, 0)

	for col, value := range s.colToValue {
		colIdx, ok := s.store.colsMapping[col]
		if !ok {
			return nil, fmt.Errorf("failed to update, unknown column name provided: %s", col)
		}

		escapedValue, err := escapeValue(col, value, s.store.colsWithFormula)
		if err != nil {
			return nil, err
		}
		if err = common.CheckIEEE754SafeInteger(escapedValue); err != nil {
			return nil, err
		}

		for _, rowIdx := range rowIndices {
			a1Range := colIdx.Name + strconv.FormatInt(rowIdx, 10)
			requests = append(requests, sheets.BatchUpdateRowsRequest{
				A1Range: common.GetA1Range(s.store.sheetName, a1Range),
				Values:  [][]interface{}{{escapedValue}},
			})
		}
	}

	return requests, nil
}

func newGoogleSheetUpdateStmt(store *GoogleSheetRowStore, colToValue map[string]interface{}) *GoogleSheetUpdateStmt {
	return &GoogleSheetUpdateStmt{
		store:        store,
		colToValue:   colToValue,
		queryBuilder: newQueryBuilder(store.colsMapping.NameMap(), ridWhereClauseInterceptor, []string{rowIdxCol}),
	}
}

// GoogleSheetDeleteStmt encapsulates information required to delete rows.
type GoogleSheetDeleteStmt struct {
	store        *GoogleSheetRowStore
	queryBuilder *queryBuilder
}

// Where specifies the condition to choose which rows are affected.
//
// It works just like the GoogleSheetSelectStmt.Where() method.
// Please read GoogleSheetSelectStmt.Where() for more details.
func (s *GoogleSheetDeleteStmt) Where(condition string, args ...interface{}) *GoogleSheetDeleteStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// Exec deletes rows matching the condition.
//
// There are 2 API calls behind the scene.
func (s *GoogleSheetDeleteStmt) Exec(ctx context.Context) error {
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

func newGoogleSheetDeleteStmt(store *GoogleSheetRowStore) *GoogleSheetDeleteStmt {
	return &GoogleSheetDeleteStmt{
		store:        store,
		queryBuilder: newQueryBuilder(store.colsMapping.NameMap(), ridWhereClauseInterceptor, []string{rowIdxCol}),
	}
}

// GoogleSheetCountStmt encapsulates information required to count the number of rows matching some conditions.
type GoogleSheetCountStmt struct {
	store        *GoogleSheetRowStore
	queryBuilder *queryBuilder
}

// Where specifies the condition to choose which rows are affected.
//
// It works just like the GoogleSheetSelectStmt.Where() method.
// Please read GoogleSheetSelectStmt.Where() for more details.
func (s *GoogleSheetCountStmt) Where(condition string, args ...interface{}) *GoogleSheetCountStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

// Exec counts the number of rows matching the provided condition.
//
// There is only 1 API call behind the scene.
func (s *GoogleSheetCountStmt) Exec(ctx context.Context) (uint64, error) {
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

func newGoogleSheetCountStmt(store *GoogleSheetRowStore) *GoogleSheetCountStmt {
	countClause := fmt.Sprintf("COUNT(%s)", rowIdxCol)
	return &GoogleSheetCountStmt{
		store:        store,
		queryBuilder: newQueryBuilder(store.colsMapping.NameMap(), ridWhereClauseInterceptor, []string{countClause}),
	}
}

func getRowIndices(ctx context.Context, store *GoogleSheetRowStore, selectStmt string) ([]int64, error) {
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

func generateRowA1Ranges(sheetName string, indices []int64) []string {
	locations := make([]string, len(indices))
	for i := range indices {
		locations[i] = common.GetA1Range(
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

func escapeValue(
	col string,
	value any,
	colsWithFormula *common.Set[string],
) (any, error) {
	if !colsWithFormula.Contains(col) {
		return common.EscapeValue(value), nil
	}

	_, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("value of column %s is not a string, but expected to contain formula", col)
	}
	return value, nil
}
