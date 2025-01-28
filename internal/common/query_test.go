package common

import (
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ridWhereClauseInterceptor(where string) string {
	if where == "" {
		return "_rid is not null"
	}
	return fmt.Sprintf("_rid is not null AND %s", where)
}

func TestGenerateQuery(t *testing.T) {
	colsMapping := models.ColsMapping{
		"_rid": {"A", 0},
		"col1": {"B", 1},
		"col2": {"C", 2},
	}

	t.Run("successful_basic", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null", result)
	})

	t.Run("unsuccessful_basic_wrong_column", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2", "col3"})
		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C, col3 where A is not null", result)
	})

	t.Run("successful_with_where", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Where("(col1 > ? AND col2 <= ?) OR (col1 != ? AND col2 == ?)", 100, true, "value", 3.14)

		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null AND (B > 100 AND C <= true ) OR (B != \"value\" AND C == 3.14 )", result)
	})

	t.Run("unsuccessful_with_where_wrong_arg_count", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Where("(col1 > ? AND col2 <= ?) OR (col1 != ? AND col2 == ?)", 100, true)

		result, err := builder.Generate()
		assert.NotNil(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("unsuccessful_with_where_unsupported_arg_type", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Where("(col1 > ? AND col2 <= ?) OR (col1 != ? AND col2 == ?)", 100, true, nil, []string{})

		result, err := builder.Generate()
		assert.NotNil(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("successful_with_limit_offset", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.Limit(10).Offset(100)

		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null offset 100 limit 10", result)
	})

	t.Run("successful_with_order_by", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
		builder.OrderBy([]models.ColumnOrderBy{{Column: "col2", OrderBy: models.OrderByDesc}, {Column: "col1", OrderBy: models.OrderByAsc}})

		result, err := builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, "select B, C where A is not null order by C DESC, B ASC", result)
	})

	t.Run("test_argument_types", func(t *testing.T) {
		builder := NewQueryBuilder(colsMapping.NameMap(), ridWhereClauseInterceptor, []string{"col1", "col2"})
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
