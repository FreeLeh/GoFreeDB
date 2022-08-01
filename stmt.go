package freeleh

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/FreeLeh/GoFreeLeh/internal/google/sheets"
	"github.com/mitchellh/mapstructure"
)

type googleSheetSelectStmt struct {
	store     *GoogleSheetRowStore
	replacer  *strings.Replacer
	columns   []string
	where     string
	whereArgs []interface{}
	orderBy   []string
	limit     uint64
	offset    uint64
	output    interface{}
}

func (s *googleSheetSelectStmt) Where(condition string, args ...interface{}) *googleSheetSelectStmt {
	s.where = condition
	s.whereArgs = args
	return s
}

func (s *googleSheetSelectStmt) OrderBy(ordering []ColumnOrderBy) *googleSheetSelectStmt {
	orderBy := make([]string, 0, len(ordering))
	for _, o := range ordering {
		orderBy = append(orderBy, o.Column+" "+string(o.OrderBy))
	}

	s.orderBy = orderBy
	return s
}

func (s *googleSheetSelectStmt) Limit(limit uint64) *googleSheetSelectStmt {
	s.limit = limit
	return s
}

func (s *googleSheetSelectStmt) Offset(offset uint64) *googleSheetSelectStmt {
	s.offset = offset
	return s
}

func (s *googleSheetSelectStmt) Exec(ctx context.Context) error {
	if err := s.ensureOutputSlice(); err != nil {
		return err
	}

	stmt, err := s.generateSelect()
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
	return mapstructure.Decode(m, s.output)
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

func (s *googleSheetSelectStmt) generateSelect() (string, error) {
	stmt := &strings.Builder{}
	stmt.WriteString("select")

	if err := s.writeCols(stmt); err != nil {
		return "", err
	}
	if err := s.writeWhere(stmt); err != nil {
		return "", err
	}
	if err := s.writeOrderBy(stmt); err != nil {
		return "", err
	}
	if err := s.writeOffset(stmt); err != nil {
		return "", err
	}
	if err := s.writeLimit(stmt); err != nil {
		return "", err
	}

	return stmt.String(), nil
}

func (s *googleSheetSelectStmt) writeCols(stmt *strings.Builder) error {
	stmt.WriteString(" ")

	translated := make([]string, 0, len(s.columns))
	for _, col := range s.columns {
		translated = append(translated, s.replacer.Replace(col))
	}

	stmt.WriteString(strings.Join(translated, ", "))
	return nil
}

func (s *googleSheetSelectStmt) writeWhere(stmt *strings.Builder) error {
	if len(s.where) == 0 {
		return nil
	}

	nArgs := strings.Count(s.where, "?")
	if nArgs != len(s.whereArgs) {
		return fmt.Errorf("number of arguments required in the 'where' clause (%d) is not the same as the number of provided arguments (%d)", nArgs, len(s.whereArgs))
	}

	where := s.replacer.Replace(s.where)
	tokens := strings.Split(where, "?")

	result := make([]string, 0)
	result = append(result, strings.TrimSpace(tokens[0]))

	for i, token := range tokens[1:] {
		arg, err := s.convertArg(s.whereArgs[i])
		if err != nil {
			return fmt.Errorf("failed converting 'where' arguments: %v, %w", arg, err)
		}
		result = append(result, arg, strings.TrimSpace(token))
	}

	stmt.WriteString(" where ")
	stmt.WriteString(strings.Join(result, " "))
	return nil
}

func (s *googleSheetSelectStmt) convertArg(arg interface{}) (string, error) {
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

func (s *googleSheetSelectStmt) writeOrderBy(stmt *strings.Builder) error {
	if len(s.orderBy) == 0 {
		return nil
	}

	stmt.WriteString(" order by ")
	result := make([]string, 0, len(s.orderBy))

	for _, o := range s.orderBy {
		result = append(result, s.replacer.Replace(o))
	}

	stmt.WriteString(strings.Join(result, ", "))
	return nil
}

func (s *googleSheetSelectStmt) writeOffset(stmt *strings.Builder) error {
	if s.offset == 0 {
		return nil
	}

	stmt.WriteString(" offset ")
	stmt.WriteString(strconv.FormatInt(int64(s.offset), 10))
	return nil
}

func (s *googleSheetSelectStmt) writeLimit(stmt *strings.Builder) error {
	if s.limit == 0 {
		return nil
	}

	stmt.WriteString(" limit ")
	stmt.WriteString(strconv.FormatInt(int64(s.limit), 10))
	return nil
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

func newGoogleSheetSelectStmt(store *GoogleSheetRowStore, output interface{}, columns []string) *googleSheetSelectStmt {
	replacements := make([]string, 0)
	for col, val := range store.colsMapping {
		replacements = append(replacements, col, val.name)
	}
	return newGoogleSheetSelectStmtWithReplacer(store, output, columns, strings.NewReplacer(replacements...))
}

func newGoogleSheetSelectStmtWithReplacer(
	store *GoogleSheetRowStore,
	output interface{},
	columns []string,
	replacer *strings.Replacer,
) *googleSheetSelectStmt {
	if len(columns) == 0 {
		columns = store.config.Columns
	}

	return &googleSheetSelectStmt{
		store:    store,
		replacer: replacer,
		columns:  columns,
		output:   output,
	}
}

type googleSheetRawInsertStmt struct {
	store *GoogleSheetRowStore
	rows  [][]interface{}
}

func (s *googleSheetRawInsertStmt) Exec(ctx context.Context) error {
	if len(s.rows) == 0 {
		return nil
	}

	_, err := s.store.wrapper.OverwriteRows(
		ctx,
		s.store.spreadsheetID,
		getA1Range(s.store.sheetName, defaultRowFullTableRange),
		s.rows,
	)
	return err
}

func newGoogleSheetRawInsertStmt(store *GoogleSheetRowStore, rows [][]interface{}) *googleSheetRawInsertStmt {
	return &googleSheetRawInsertStmt{
		store: store,
		rows:  rows,
	}
}

type googleSheetInsertStmt struct {
	store *GoogleSheetRowStore
}

type googleSheetUpdateStmt struct {
	store      *GoogleSheetRowStore
	colToValue map[string]interface{}
	where      string
	whereArgs  []interface{}
}

func (s *googleSheetUpdateStmt) Where(condition string, args ...interface{}) *googleSheetUpdateStmt {
	s.where = condition
	s.whereArgs = args
	return s
}

func (s *googleSheetUpdateStmt) Exec(ctx context.Context) error {
	// The first _ts IS NOT NULL is necessary to ensure we are just updating rows that are non-empty.
	// This is required for UPDATE without WHERE clause (otherwise it will see every row as update target).
	selectStmt, err := generateSelectQuery(s.store, s.generateWhere(), s.whereArgs)
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

func (s *googleSheetUpdateStmt) generateWhere() string {
	if s.where == "" {
		return fmt.Sprintf(rowUpdateModifyWhereEmptyTemplate, rowTsCol)
	}
	return fmt.Sprintf(rowUpdateModifyWhereNonEmptyTemplate, rowTsCol, s.where)
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
		store:      store,
		colToValue: colToValue,
	}
}

type googleSheetDeleteStmt struct {
	store     *GoogleSheetRowStore
	where     string
	whereArgs []interface{}
}

func (s *googleSheetDeleteStmt) Where(condition string, args ...interface{}) *googleSheetDeleteStmt {
	s.where = condition
	s.whereArgs = args
	return s
}

func (s *googleSheetDeleteStmt) Exec(ctx context.Context) error {
	selectStmt, err := generateSelectQuery(s.store, s.where, s.whereArgs)
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
	return &googleSheetDeleteStmt{store: store}
}

func generateSelectQuery(store *GoogleSheetRowStore, where string, whereArgs []interface{}) (string, error) {
	replacements := make([]string, 0)
	for col, val := range store.colsMapping {
		replacements = append(replacements, col, "Col"+strconv.FormatInt(int64(val.idx+1), 10))
	}

	col := []string{"Col" + strconv.FormatInt(int64(maxColumn+1), 10)}
	return newGoogleSheetSelectStmtWithReplacer(store, nil, col, strings.NewReplacer(replacements...)).
		Where(where, whereArgs...).
		generateSelect()
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
