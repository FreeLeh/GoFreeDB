package lark

import (
	"context"
	"go.uber.org/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type person struct {
	Name string `db:"name,omitempty"`
	Age  int64  `db:"age,omitempty"`
	DOB  string `db:"dob,omitempty"`
}

func TestSelectStmt_AllColumns(t *testing.T) {
	store := &SheetRowStore{
		colsMapping: models.ColsMapping{
			rowIdxCol: {"A", 0},
			"col1":    {"B", 1},
			"col2":    {"C", 2},
		},
		config: SheetRowStoreConfig{Columns: []string{"col1", "col2"}},
	}
	stmt := newSheetSelectStmt(store, nil, []string{})

	result, err := stmt.queryBuilder.Generate()
	assert.Nil(t, err)
	assert.Equal(t, "select B, C where A is not null", result)
}

func TestSelectStmt_Exec(t *testing.T) {
	t.Run("non_slice_output", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wrapper := NewMocksheetsWrapper(ctrl)
		store := &SheetRowStore{
			wrapper: wrapper,
			colsMapping: map[string]models.ColIdx{
				rowIdxCol: {"A", 0},
				"col1":    {"B", 1},
				"col2":    {"C", 2},
			},
		}
		o := 0
		stmt := newSheetSelectStmt(store, &o, []string{"col1", "col2"})

		assert.NotNil(t, stmt.Exec(context.Background()))
	})

	t.Run("non_pointer_to_slice_output", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wrapper := NewMocksheetsWrapper(ctrl)
		store := &SheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]models.ColIdx{rowIdxCol: {"A", 0}, "col1": {"B", 1}, "col2": {"C", 2}},
		}
		var o []int
		stmt := newSheetSelectStmt(store, o, []string{"col1", "col2"})

		assert.NotNil(t, stmt.Exec(context.Background()))
	})

	t.Run("nil_output", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wrapper := NewMocksheetsWrapper(ctrl)
		store := &SheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]models.ColIdx{rowIdxCol: {"A", 0}, "col1": {"B", 1}, "col2": {"C", 2}},
		}
		stmt := newSheetSelectStmt(store, nil, []string{"col1", "col2"})

		assert.NotNil(t, stmt.Exec(context.Background()))
	})

	t.Run("has_query_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wrapper := NewMocksheetsWrapper(ctrl)
		wrapper.EXPECT().QueryRows(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(QueryRowsResult{}, assert.AnError).AnyTimes()
		store := &SheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]models.ColIdx{rowIdxCol: {"A", 0}, "col1": {"B", 1}, "col2": {"C", 2}},
		}
		var out []int
		stmt := newSheetSelectStmt(store, &out, []string{"col1", "col2"})

		err := stmt.Exec(context.Background())
		assert.NotNil(t, err)
	})

	t.Run("successful", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wrapper := NewMocksheetsWrapper(ctrl)

		gomock.InOrder(
			wrapper.EXPECT().QueryRows(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(QueryRowsResult{Rows: [][]interface{}{
				{2.0},
			}}, nil).Times(1),
			wrapper.EXPECT().QueryRows(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(QueryRowsResult{Rows: [][]interface{}{
				{10, "17-01-2001"},
				{11, "18-01-2000"},
			}}, nil).Times(1),
		)

		wrapper.EXPECT().
			BatchUpdateRows(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]BatchUpdateRowsResult{}, nil).
			AnyTimes()

		store := &SheetRowStore{
			wrapper:     wrapper,
			colsMapping: map[string]models.ColIdx{rowIdxCol: {"A", 0}, "name": {"B", 1}, "age": {"C", 2}, "dob": {"D", 3}},
			config: SheetRowStoreConfig{
				Columns: []string{"name", "age", "dob"},
			},
		}
		var out []person
		stmt := newSheetSelectStmt(store, &out, []string{"age", "dob"})

		expected := []person{
			{Age: 10, DOB: "17-01-2001"},
			{Age: 11, DOB: "18-01-2000"},
		}
		err := stmt.Exec(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("successful_select_all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wrapper := NewMocksheetsWrapper(ctrl)

		gomock.InOrder(
			wrapper.EXPECT().QueryRows(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(QueryRowsResult{Rows: [][]interface{}{
				{2.0},
			}}, nil).Times(1),
			wrapper.EXPECT().QueryRows(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(QueryRowsResult{Rows: [][]interface{}{
				{"name1", 10, "17-01-2001"},
				{"name2", 11, "18-01-2000"},
			}}, nil).Times(1),
		)

		wrapper.EXPECT().
			BatchUpdateRows(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]BatchUpdateRowsResult{}, nil).
			AnyTimes()

		store := &SheetRowStore{
			wrapper: wrapper,
			colsMapping: map[string]models.ColIdx{
				rowIdxCol: {"A", 0},
				"name":    {"B", 1},
				"age":     {"C", 2},
				"dob":     {"D", 3},
			},
			colsWithFormula: common.NewSet([]string{"name"}),
			config: SheetRowStoreConfig{
				Columns: []string{"name", "age", "dob"},
			},
		}
		var out []person
		stmt := newSheetSelectStmt(store, &out, []string{})

		expected := []person{
			{Name: "name1", Age: 10, DOB: "17-01-2001"},
			{Name: "name2", Age: 11, DOB: "18-01-2000"},
		}
		err := stmt.Exec(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, expected, out)
	})
}

func TestSheetInsertStmt_convertRowToSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	wrapper := NewMocksheetsWrapper(ctrl)
	store := &SheetRowStore{
		wrapper: wrapper,
		colsMapping: map[string]models.ColIdx{
			rowIdxCol: {"A", 0},
			"name":    {"B", 1},
			"age":     {"C", 2},
			"dob":     {"D", 3},
		},
		colsWithFormula: common.NewSet([]string{"name"}),
		config: SheetRowStoreConfig{
			Columns: []string{"name", "age", "dob"},
		},
	}

	t.Run("non_struct", func(t *testing.T) {
		stmt := newSheetInsertStmt(store, nil)

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
		stmt := newSheetInsertStmt(store, nil)

		result, err := stmt.convertRowToSlice(person{Name: "blah", Age: 10, DOB: "2021"})
		assert.Equal(t, []interface{}{convertToFormula(rowIdxFormula), "blah", int64(10), "2021"}, result)
		assert.Nil(t, err)

		result, err = stmt.convertRowToSlice(&person{Name: "blah", Age: 10, DOB: "2021"})
		assert.Equal(t, []interface{}{convertToFormula(rowIdxFormula), "blah", int64(10), "2021"}, result)
		assert.Nil(t, err)

		result, err = stmt.convertRowToSlice(person{Name: "blah", DOB: "2021"})
		assert.Equal(t, []interface{}{convertToFormula(rowIdxFormula), "blah", nil, "2021"}, result)
		assert.Nil(t, err)

		type dummy struct {
			Name string `db:"name"`
		}

		result, err = stmt.convertRowToSlice(dummy{Name: "blah"})
		assert.Equal(t, []interface{}{convertToFormula(rowIdxFormula), "blah", nil, nil}, result)
		assert.Nil(t, err)
	})

	t.Run("ieee754_safe_integers", func(t *testing.T) {
		stmt := newSheetInsertStmt(store, nil)

		result, err := stmt.convertRowToSlice(person{Name: "blah", Age: 9007199254740992, DOB: "2021"})
		assert.Equal(t, []interface{}{convertToFormula(rowIdxFormula), "blah", int64(9007199254740992), "2021"}, result)
		assert.Nil(t, err)

		result, err = stmt.convertRowToSlice(person{Name: "blah", Age: 9007199254740993, DOB: "2021"})
		assert.Nil(t, result)
		assert.NotNil(t, err)
	})
}

func TestGoogleSheetUpdateStmt_generateBatchUpdateRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	wrapper := NewMocksheetsWrapper(ctrl)
	store := &SheetRowStore{
		wrapper:   wrapper,
		sheetName: "sheet1",
		colsMapping: map[string]models.ColIdx{
			rowIdxCol: {"A", 0},
			"name":    {"B", 1},
			"age":     {"C", 2},
			"dob":     {"D", 3},
		},
		colsWithFormula: common.NewSet([]string{"name"}),
		config: SheetRowStoreConfig{
			Columns: []string{"name", "age", "dob"},
		},
	}

	t.Run("successful", func(t *testing.T) {
		stmt := newSheetUpdateStmt(
			store,
			map[string]interface{}{
				"name": "name1",
				"age":  int64(100),
				"dob":  "hello",
			},
		)

		requests, err := stmt.generateBatchUpdateRequests([]int64{1, 2})
		expected := []BatchUpdateRowsRequest{
			{
				A1Range: models.NewA1Range(store.sheetID, "B1:B1"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "B2:B2"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "C1:C1"),
				Values:  [][]interface{}{{int64(100)}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "C2:C2"),
				Values:  [][]interface{}{{int64(100)}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "D1:D1"),
				Values:  [][]interface{}{{"hello"}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "D2:D2"),
				Values:  [][]interface{}{{"hello"}},
			},
		}

		assert.ElementsMatch(t, expected, requests)
		assert.Nil(t, err)
	})

	t.Run("ieee754_safe_integers_successful", func(t *testing.T) {
		stmt := newSheetUpdateStmt(store, map[string]interface{}{
			"name": "name1",
			"age":  int64(9007199254740992),
		})

		requests, err := stmt.generateBatchUpdateRequests([]int64{1, 2})
		expected := []BatchUpdateRowsRequest{
			{
				A1Range: models.NewA1Range(store.sheetID, "B1:B1"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "B2:B2"),
				Values:  [][]interface{}{{"name1"}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "C1:C1"),
				Values:  [][]interface{}{{int64(9007199254740992)}},
			},
			{
				A1Range: models.NewA1Range(store.sheetID, "C2:C2"),
				Values:  [][]interface{}{{int64(9007199254740992)}},
			},
		}

		assert.ElementsMatch(t, expected, requests)
		assert.Nil(t, err)
	})

	t.Run("ieee754_safe_integers_unsuccessful", func(t *testing.T) {
		stmt := newSheetUpdateStmt(
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
