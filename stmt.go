package freeleh

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/FreeLeh/GoFreeLeh/internal/google/sheets"
)

type queryBuilder struct {
	replacer  *strings.Replacer
	columns   []string
	where     string
	whereArgs []interface{}
	orderBy   []string
	limit     uint64
	offset    uint64
}

func (q *queryBuilder) Where(condition string, args ...interface{}) *queryBuilder {
	q.where = condition
	q.whereArgs = args
	return q
}

func (q *queryBuilder) OrderBy(ordering []ColumnOrderBy) *queryBuilder {
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
	if len(q.where) == 0 {
		return nil
	}

	nArgs := strings.Count(q.where, "?")
	if nArgs != len(q.whereArgs) {
		return fmt.Errorf("number of arguments required in the 'where' clause (%d) is not the same as the number of provided arguments (%d)", nArgs, len(q.whereArgs))
	}

	where := q.replacer.Replace(q.where)
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
	case float32:
		return strconv.FormatFloat(float64(converted), 'f', -1, 64), nil
	case float64:
		return strconv.FormatFloat(converted, 'f', -1, 64), nil
	case string:
		cleaned := strings.ToLower(strings.TrimSpace(converted))
		if googleSheetSelectStmtStringKeyword.MatchString(cleaned) {
			return converted, nil
		}
		return "'" + converted + "'", nil
	case []byte:
		return strconv.Quote(string(converted)), nil
	case bool:
		return strconv.FormatBool(converted), nil
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

func newQueryBuilder(colReplacements map[string]string, colSelected []string) *queryBuilder {
	replacements := make([]string, 0, 2*len(colReplacements))
	for col, repl := range colReplacements {
		replacements = append(replacements, col, repl)
	}

	return &queryBuilder{replacer: strings.NewReplacer(replacements...), columns: colSelected}
}

type googleSheetSelectStmt struct {
	store        *GoogleSheetRowStore
	columns      []string
	queryBuilder *queryBuilder
	output       interface{}
}

func (s *googleSheetSelectStmt) Where(condition string, args ...interface{}) *googleSheetSelectStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

func (s *googleSheetSelectStmt) OrderBy(ordering []ColumnOrderBy) *googleSheetSelectStmt {
	s.queryBuilder.OrderBy(ordering)
	return s
}

func (s *googleSheetSelectStmt) Limit(limit uint64) *googleSheetSelectStmt {
	s.queryBuilder.Limit(limit)
	return s
}

func (s *googleSheetSelectStmt) Offset(offset uint64) *googleSheetSelectStmt {
	s.queryBuilder.Offset(offset)
	return s
}

func (s *googleSheetSelectStmt) Exec(ctx context.Context) error {
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
	return mapstructureDecode(m, s.output)
}

func (s *googleSheetSelectStmt) buildQueryResultMap(original sheets.QueryRowsResult) []map[string]interface{} {
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

func (s *googleSheetSelectStmt) ensureOutputSlice() error {
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

func newGoogleSheetSelectStmt(store *GoogleSheetRowStore, output interface{}, columns []string) *googleSheetSelectStmt {
	if len(columns) == 0 {
		columns = store.config.Columns
	}

	return &googleSheetSelectStmt{
		store:        store,
		columns:      columns,
		queryBuilder: newQueryBuilder(store.colsMapping.NameMap(), columns),
		output:       output,
	}
}

type googleSheetInsertStmt struct {
	store *GoogleSheetRowStore
	rows  []interface{}
}

func (s *googleSheetInsertStmt) convertRowToSlice(row interface{}, curTs int64) ([]interface{}, error) {
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
	if err := mapstructureDecode(row, &output); err != nil {
		return nil, err
	}

	result := make([]interface{}, len(s.store.colsMapping))
	for key, value := range output {
		if colIdx, ok := s.store.colsMapping[key]; ok {
			result[colIdx.idx] = value
		}
	}

	// Insert the _ts value.
	result[len(s.store.colsMapping)-1] = curTs
	return result, nil
}

func (s *googleSheetInsertStmt) Exec(ctx context.Context) error {
	if len(s.rows) == 0 {
		return nil
	}

	convertedRows := make([][]interface{}, 0, len(s.rows))
	curTs := currentTimeMs()

	for _, row := range s.rows {
		r, err := s.convertRowToSlice(row, curTs)
		if err != nil {
			return fmt.Errorf("cannot execute google sheet insert statement due to row conversion error: %w", err)
		}
		convertedRows = append(convertedRows, r)
	}

	_, err := s.store.wrapper.OverwriteRows(
		ctx,
		s.store.spreadsheetID,
		getA1Range(s.store.sheetName, defaultRowFullTableRange),
		convertedRows,
	)
	return err
}

func newGoogleSheetInsertStmt(store *GoogleSheetRowStore, rows []interface{}) *googleSheetInsertStmt {
	return &googleSheetInsertStmt{
		store: store,
		rows:  rows,
	}
}

type googleSheetUpdateStmt struct {
	store        *GoogleSheetRowStore
	colToValue   map[string]interface{}
	queryBuilder *queryBuilder
}

func (s *googleSheetUpdateStmt) Where(condition string, args ...interface{}) *googleSheetUpdateStmt {
	// The first condition `_ts IS NOT NULL` is necessary to ensure we are just updating rows that are non-empty.
	// This is required for UPDATE without WHERE clause (otherwise it will see every row as update target).
	if condition == "" {
		s.queryBuilder.Where(fmt.Sprintf(rowUpdateModifyWhereEmptyTemplate, rowTsCol), args...)
	} else {
		s.queryBuilder.Where(fmt.Sprintf(rowUpdateModifyWhereNonEmptyTemplate, rowTsCol, condition), args...)
	}
	return s
}

func (s *googleSheetUpdateStmt) Exec(ctx context.Context) error {
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

func (s *googleSheetUpdateStmt) generateBatchUpdateRequests(rowIndices []int) ([]sheets.BatchUpdateRowsRequest, error) {
	requests := make([]sheets.BatchUpdateRowsRequest, 0)

	for col, value := range s.colToValue {
		colIdx, ok := s.store.colsMapping[col]
		if !ok {
			return nil, fmt.Errorf("failed to update, unknown column name provided: %s", col)
		}

		for _, rowIdx := range rowIndices {
			a1Range := colIdx.name + strconv.FormatInt(int64(rowIdx), 10)
			requests = append(requests, sheets.BatchUpdateRowsRequest{
				A1Range: getA1Range(s.store.sheetName, a1Range),
				Values:  [][]interface{}{{value}},
			})
		}
	}

	return requests, nil
}

func newGoogleSheetUpdateStmt(store *GoogleSheetRowStore, colToValue map[string]interface{}) *googleSheetUpdateStmt {
	return &googleSheetUpdateStmt{
		store:        store,
		colToValue:   colToValue,
		queryBuilder: newQueryBuilder(store.colsMapping.ColIdxNameMap(), []string{lastColIdxName}),
	}
}

type googleSheetDeleteStmt struct {
	store        *GoogleSheetRowStore
	queryBuilder *queryBuilder
}

func (s *googleSheetDeleteStmt) Where(condition string, args ...interface{}) *googleSheetDeleteStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

func (s *googleSheetDeleteStmt) Exec(ctx context.Context) error {
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

func newGoogleSheetDeleteStmt(store *GoogleSheetRowStore) *googleSheetDeleteStmt {
	return &googleSheetDeleteStmt{
		store:        store,
		queryBuilder: newQueryBuilder(store.colsMapping.ColIdxNameMap(), []string{lastColIdxName}),
	}
}

type googleSheetCountStmt struct {
	store        *GoogleSheetRowStore
	queryBuilder *queryBuilder
}

func (s *googleSheetCountStmt) Where(condition string, args ...interface{}) *googleSheetCountStmt {
	s.queryBuilder.Where(condition, args...)
	return s
}

func (s *googleSheetCountStmt) Exec(ctx context.Context) (uint64, error) {
	selectStmt, err := s.queryBuilder.Generate()
	if err != nil {
		return 0, err
	}

	formula := fmt.Sprintf(
		rwoCountQueryTemplate,
		getA1Range(s.store.sheetName, defaultRowFullTableRange),
		selectStmt,
	)

	result, err := s.store.wrapper.UpdateRows(
		ctx,
		s.store.spreadsheetID,
		s.store.scratchpadLocation.Original,
		[][]interface{}{{formula}},
	)
	if err != nil {
		return 0, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return 0, fmt.Errorf("error retrieving row indices to delete: %+v", result)
	}

	raw := result.UpdatedValues[0][0].(string)
	if raw == naValue || raw == errorValue || raw == "" {
		return 0, fmt.Errorf("error retrieving row indices to delete: %s", raw)
	}
	return strconv.ParseUint(raw, 10, 64)
}

func newGoogleSheetCountStmt(store *GoogleSheetRowStore) *googleSheetCountStmt {
	return &googleSheetCountStmt{
		store:        store,
		queryBuilder: newQueryBuilder(store.colsMapping.NameMap(), []string{rowTsCol}),
	}
}

func getRowIndices(ctx context.Context, store *GoogleSheetRowStore, selectStmt string) ([]int, error) {
	formula := fmt.Sprintf(
		rowGetIndicesQueryTemplate,
		getA1Range(store.sheetName, defaultRowFullTableRange),
		getA1Range(store.sheetName, defaultRowFullTableRange),
		selectStmt,
	)

	result, err := store.wrapper.UpdateRows(
		ctx,
		store.spreadsheetID,
		store.scratchpadLocation.Original,
		[][]interface{}{{formula}},
	)
	if err != nil {
		return nil, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return nil, fmt.Errorf("error retrieving row indices to delete: %+v", result)
	}

	raw := result.UpdatedValues[0][0].(string)
	if raw == naValue {
		return []int{}, nil
	}
	if raw == errorValue || raw == "" {
		return nil, fmt.Errorf("error retrieving row indices to delete: %s", raw)
	}

	rowIndicesStr := strings.Split(raw, ",")
	rowIndices := make([]int, len(rowIndicesStr))

	for i := range rowIndicesStr {
		idx, err := strconv.ParseInt(rowIndicesStr[i], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error converting row indices to delete: %w", err)
		}
		rowIndices[i] = int(idx)
	}

	return rowIndices, nil
}

func generateRowA1Ranges(sheetName string, indices []int) []string {
	locations := make([]string, len(indices))
	for i := range indices {
		locations[i] = getA1Range(
			sheetName,
			fmt.Sprintf(rowDeleteRangeTemplate, indices[i], indices[i]),
		)
	}
	return locations
}
