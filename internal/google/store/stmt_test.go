package store

import (
	"context"
	"errors"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"testing"

	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
	"github.com/stretchr/testify/assert"
)

type person struct {
	Name string `db:"name,omitempty"`
	Age  int64  `db:"age,omitempty"`
	DOB  string `db:"dob,omitempty"`
}

func TestGenerateQuery(t *testing.T) {
	colsMapping := common.ColsMapping{
		rowIdxCol: {"A", 0},
		"col1":    {"B", 1},
		"col2":    {"C", 2},
	}

	t.Run("successful_basic", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null", result)
	})

	t.Run("unsuccessful_basic_wrong_column", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2", "col3"})
		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C, col3 where A is not null", result)
	})

	t.Run("successful_with_where", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Where("(col1 > ? AND col2 <= ?) OR (col1 != ? AND col2 == ?)", 100, true, "value", 3.14)

		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null AND (B > 100 AND C <= true ) OR (B != \"value\" AND C == 3.14 )", result)
	})

	t.Run("unsuccessful_with_where_wrong_arg_count", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Where("(col1 > ? AND col2 <= ?) OR (col1 != ? AND col2 == ?)", 100, true)

		result, err := builder.Generate()
		assert.NotNil(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("unsuccessful_with_where_unsupported_arg_type", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Where("(col1 > ? AND col2 <= ?) OR (col1 != ? AND col2 == ?)", 100, true, nil, []string{})

		result, err := builder.Generate()
		assert.NotNil(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("successful_with_limit_offset", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Limit(10).Offset(100)

		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null offset 100 limit 10", result)
	})

	t.Run("successful_with_order_by", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.OrderBy([]models.ColumnOrderBy{{Column: "col2", OrderBy: models.OrderByDesc}, {Column: "col1", OrderBy: models.OrderByAsc}})

		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null order by C DESC, B ASC", result)
	})

	t.Run("test_argument_types", func(t *testing.T) {
		builder := newQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		tc := []struct {
			input  interface{}
			output string
			err    error
		}{
			{
				input:  1,
				output: "1",
				err:    nil,
			},
			{
				input:  int8(1),
				output: "1",
				err:    nil,
			},
			{
				input:  int16(1),
				output: "1",
				err:    nil,
			},
			{
				input:  int32(1),
				output: "1",
				err:    nil,
			},
			{
				input:  int64(1),
				output: "1",
				err:    nil,
			},
			{
				input:  uint(1),
				output: "1",
				err:    nil,
			},
			{
				input:  uint8(1),
				output: "1",
				err:    nil,
			},
			{
				input:  uint16(1),
				output: "1",
				err:    nil,
			},
			{
				input:  uint32(1),
				output: "1",
				err:    nil,
			},
			{
				input:  uint64(1),
				output: "1",
				err:    nil,
			},
			{
				input:  float32(1.5),
				output: "1.5",
				err:    nil,
			},
			{
				input:  1.5,
				output: "1.5",
				err:    nil,
			},
			{
				input:  "something",
				output: "\"something\"",
				err:    nil,
			},
			{
				input:  "date",
				output: "date",
				err:    nil,
			},
			{
				input:  "datetime",
				output: "datetime",
				err:    nil,
			},
			{
				input:  "timeofday",
				output: "timeofday",
				err:    nil,
			},
			{
				input:  true,
				output: "true",
				err:    nil,
			},
			{
				input:  []byte("something"),
				output: "\"something\"",
				err:    nil,
			},
		}

		for _, c := range tc {
			result, err := builder.convertArg(c.input)
			assert.Equal(t, c.output, result)
			assert.Equal(t, c.err, err)
		}
	})
}

func TestSelectStmt_AllColumns(t *testing.T) {
	store := &GoogleSheetRowStore{
		colsMapping: common.ColsMapping{
			rowIdxCol: {"A", 0},
			"col1":    {"B", 1},
			"col2":    {"C", 2},
		},
		config: GoogleSheetRowStoreConfig{Columns: []string{"col1", "col2"}},
	}
	stmt := newGoogleSheetSelectStmt(store, nil, []string{})

	result, err := stmt.queryBuilder.Generate()
	assert.Nil(t, err)
	assert.Equal(t, "select B, C where A is not null", result)
}

func TestSelectStmt_Exec(t *testing.T) {
	t.Run("non_slice_output", func(t *testing.T) {
		wrapper := &sheets.MockWrapper{}
		store := &GoogleSheetRowStore{
			wrapper: wrapper,
			colsMapping: map[string]common.ColIdx{
				rowIdxCol: {"A", 0},
				"col1":    {"B", 1},
				"col2":    {"C", 2},
			},
		}
		o := 0
		stmt := newGoogleSheetSelectStmt(store, &o, []string{"col1", "col2"})

		assert.NotNil(t, stmt.Exec(context.Background()))
	})

	t.Run("non_pointer_to_slice_output", func(t *testing.T) {
		wrapper := &sheets.MockWrapper{}
		store := &GoogleSheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]common.ColIdx{rowIdxCol: {"A", 0}, "col1": {"B", 1}, "col2": {"C", 2}},
		}
		var o []int
		stmt := newGoogleSheetSelectStmt(store, o, []string{"col1", "col2"})

		assert.NotNil(t, stmt.Exec(context.Background()))
	})

	t.Run("nil_output", func(t *testing.T) {
		wrapper := &sheets.MockWrapper{}
		store := &GoogleSheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]common.ColIdx{rowIdxCol: {"A", 0}, "col1": {"B", 1}, "col2": {"C", 2}},
		}
		stmt := newGoogleSheetSelectStmt(store, nil, []string{"col1", "col2"})

		assert.NotNil(t, stmt.Exec(context.Background()))
	})

	t.Run("has_query_error", func(t *testing.T) {
		wrapper := &sheets.MockWrapper{QueryRowsError: errors.New("some error")}
		store := &GoogleSheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]common.ColIdx{rowIdxCol: {"A", 0}, "col1": {"B", 1}, "col2": {"C", 2}},
		}
		var out []int
		stmt := newGoogleSheetSelectStmt(store, &out, []string{"col1", "col2"})

		err := stmt.Exec(context.Background())
		assert.NotNil(t, err)
	})

	t.Run("successful", func(t *testing.T) {
		wrapper := &sheets.MockWrapper{QueryRowsResult: sheets.QueryRowsResult{Rows: [][]interface{}{
			{10, "17-01-2001"},
			{11, "18-01-2000"},
		}}}
		store := &GoogleSheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]common.ColIdx{rowIdxCol: {"A", 0}, "name": {"B", 1}, "age": {"C", 2}, "dob": {"D", 3}},
			config: GoogleSheetRowStoreConfig{
				Columns: []string{"name", "age", "dob"},
			},
		}
		var out []person
		stmt := newGoogleSheetSelectStmt(store, &out, []string{"age", "dob"})

		expected := []person{
			{Age: 10, DOB: "17-01-2001"},
			{Age: 11, DOB: "18-01-2000"},
		}
		err := stmt.Exec(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("successful_select_all", func(t *testing.T) {
		wrapper := &sheets.MockWrapper{QueryRowsResult: sheets.QueryRowsResult{Rows: [][]interface{}{
			{"name1", 10, "17-01-2001"},
			{"name2", 11, "18-01-2000"},
		}}}
		store := &GoogleSheetRowStore{
			wrapper: wrapper,
			colsMapping: map[string]common.ColIdx{
				rowIdxCol: {"A", 0},
				"name":    {"B", 1},
				"age":     {"C", 2},
				"dob":     {"D", 3},
			},
			colsWithFormula: common.NewSet([]string{"name"}),
			config: GoogleSheetRowStoreConfig{
				Columns:            []string{"name", "age", "dob"},
				ColumnsWithFormula: []string{"name"}},
		}
		var out []person
		stmt := newGoogleSheetSelectStmt(store, &out, []string{})

		expected := []person{
			{Name: "name1", Age: 10, DOB: "17-01-2001"},
			{Name: "name2", Age: 11, DOB: "18-01-2000"},
		}
		err := stmt.Exec(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, expected, out)
	})
}

func TestGoogleSheetInsertStmt_convertRowToSlice(t *testing.T) {
	wrapper := &sheets.MockWrapper{}
	store := &GoogleSheetRowStore{
		wrapper: wrapper,
		colsMapping: map[string]common.ColIdx{
			rowIdxCol: {"A", 0},
			"name":    {"B", 1},
			"age":     {"C", 2},
			"dob":     {"D", 3},
		},
		colsWithFormula: common.NewSet([]string{"name"}),
		config: GoogleSheetRowStoreConfig{
			Columns:            []string{"name", "age", "dob"},
			ColumnsWithFormula: []string{"name"},
		},
	}

	t.Run("non_struct", func(t *testing.T) {
		stmt := newGoogleSheetInsertStmt(store, nil)

		result, err := stmt.convertRowToSlice(nil)
		assert.Nil(t, result)
		assert.NotNil(t, err)

		result, err = stmt.convertRowToSlice(1)
		assert.Nil(t, result)
		assert.NotNil(t, err)

		result, err = stmt.convertRowToSlice("1")
		assert.Nil(t, result)
		assert.NotNil(t, err)

		result, err = stmt.convertRowToSlice([]int{1, 2, 3})
		assert.Nil(t, result)
		assert.NotNil(t, err)
	})

	t.Run("struct", func(t *testing.T) {
		stmt := newGoogleSheetInsertStmt(store, nil)

		result, err := stmt.convertRowToSlice(person{Name: "blah", Age: 10, DOB: "2021"})
		assert.Equal(t, []interface{}{rowIdxFormula, "blah", int64(10), "'2021"}, result)
		assert.Nil(t, err)

		result, err = stmt.convertRowToSlice(&person{Name: "blah", Age: 10, DOB: "2021"})
		assert.Equal(t, []interface{}{rowIdxFormula, "blah", int64(10), "'2021"}, result)
		assert.Nil(t, err)

		result, err = stmt.convertRowToSlice(person{Name: "blah", DOB: "2021"})
		assert.Equal(t, []interface{}{rowIdxFormula, "blah", nil, "'2021"}, result)
		assert.Nil(t, err)

		type dummy struct {
			Name string `db:"name"`
		}

		result, err = stmt.convertRowToSlice(dummy{Name: "blah"})
		assert.Equal(t, []interface{}{rowIdxFormula, "blah", nil, nil}, result)
		assert.Nil(t, err)
	})

	t.Run("ieee754_safe_integers", func(t *testing.T) {
		stmt := newGoogleSheetInsertStmt(store, nil)

		result, err := stmt.convertRowToSlice(person{Name: "blah", Age: 9007199254740992, DOB: "2021"})
		assert.Equal(t, []interface{}{rowIdxFormula, "blah", int64(9007199254740992), "'2021"}, result)
		assert.Nil(t, err)

		result, err = stmt.convertRowToSlice(person{Name: "blah", Age: 9007199254740993, DOB: "2021"})
		assert.Nil(t, result)
		assert.NotNil(t, err)
	})
}

func TestGoogleSheetUpdateStmt_generateBatchUpdateRequests(t *testing.T) {
	wrapper := &sheets.MockWrapper{}
	store := &GoogleSheetRowStore{
		wrapper:   wrapper,
		sheetName: "sheet1",
		colsMapping: map[string]common.ColIdx{
			rowIdxCol: {"A", 0},
			"name":    {"B", 1},
			"age":     {"C", 2},
			"dob":     {"D", 3},
		},
		colsWithFormula: common.NewSet([]string{"name"}),
		config: GoogleSheetRowStoreConfig{
			Columns:            []string{"name", "age", "dob"},
			ColumnsWithFormula: []string{"name"},
		},
	}

	t.Run("successful", func(t *testing.T) {
		stmt := newGoogleSheetUpdateStmt(
			store,
			map[string]interface{}{
				"name": "name1",
				"age":  int64(100),
				"dob":  "hello",
			},
		)

		requests, err := stmt.generateBatchUpdateRequests([]int64{1, 2})
		expected := []sheets.BatchUpdateRowsRequest{
			{
				A1Range: common.GetA1Range(store.sheetName, "B1"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "B2"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "C1"),
				Values:  [][]interface{}{{int64(100)}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "C2"),
				Values:  [][]interface{}{{int64(100)}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "D1"),
				Values:  [][]interface{}{{"'hello"}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "D2"),
				Values:  [][]interface{}{{"'hello"}},
			},
		}

		assert.ElementsMatch(t, expected, requests)
		assert.Nil(t, err)
	})

	t.Run("ieee754_safe_integers_successful", func(t *testing.T) {
		stmt := newGoogleSheetUpdateStmt(store, map[string]interface{}{
			"name": "name1",
			"age":  int64(9007199254740992),
		})

		requests, err := stmt.generateBatchUpdateRequests([]int64{1, 2})
		expected := []sheets.BatchUpdateRowsRequest{
			{
				A1Range: common.GetA1Range(store.sheetName, "B1"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "B2"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "C1"),
				Values:  [][]interface{}{{int64(9007199254740992)}},
			},
			{
				A1Range: common.GetA1Range(store.sheetName, "C2"),
				Values:  [][]interface{}{{int64(9007199254740992)}},
			},
		}

		assert.ElementsMatch(t, expected, requests)
		assert.Nil(t, err)
	})

	t.Run("ieee754_safe_integers_unsuccessful", func(t *testing.T) {
		stmt := newGoogleSheetUpdateStmt(
			store,
			map[string]interface{}{
				"name": "name1",
				"age":  int64(9007199254740993),
			},
		)

		requests, err := stmt.generateBatchUpdateRequests([]int64{1, 2})
		assert.Nil(t, requests)
		assert.NotNil(t, err)
	})
}

func TestEscapeValue(t *testing.T) {
	t.Run("not in cols with formula", func(t *testing.T) {
		value, err := escapeValue("A", 123, common.NewSet([]string{"B"}))
		assert.Nil(t, err)
		assert.Equal(t, 123, value)

		value, err = escapeValue("A", "123", common.NewSet([]string{"B"}))
		assert.Nil(t, err)
		assert.Equal(t, "'123", value)
	})

	t.Run("in cols with formula, but not string", func(t *testing.T) {
		value, err := escapeValue("A", 123, common.NewSet([]string{"A"}))
		assert.NotNil(t, err)
		assert.Equal(t, nil, value)
	})

	t.Run("in cols with formula, but string", func(t *testing.T) {
		value, err := escapeValue("A", "123", common.NewSet([]string{"A"}))
		assert.Nil(t, err)
		assert.Equal(t, "123", value)
	})
}
