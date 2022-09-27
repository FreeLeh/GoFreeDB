package sheets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestA1Range(t *testing.T) {
	tc := []struct {
		name      string
		input     string
		sheetName string
		fromCell  string
		toCell    string
	}{
		{
			name:      "no_sheet_name_single_range",
			input:     "A1",
			sheetName: "",
			fromCell:  "A1",
			toCell:    "A1",
		},
		{
			name:      "no_sheet_name_multiple_range",
			input:     "A1:A2",
			sheetName: "",
			fromCell:  "A1",
			toCell:    "A2",
		},
		{
			name:      "has_sheet_name_single_range",
			input:     "Sheet1!A1",
			sheetName: "Sheet1",
			fromCell:  "A1",
			toCell:    "A1",
		},
		{
			name:      "has_sheet_name_multiple_range",
			input:     "Sheet1!A1:A2",
			sheetName: "Sheet1",
			fromCell:  "A1",
			toCell:    "A2",
		},
		{
			name:      "empty_input",
			input:     "",
			sheetName: "",
			fromCell:  "",
			toCell:    "",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			a1 := NewA1Range(c.input)
			assert.Equal(t, a1.Original, c.input, "A1Range original should have the same value as the input")
			assert.Equal(t, a1.SheetName, c.sheetName, "wrong sheet name")
			assert.Equal(t, a1.FromCell, c.fromCell, "wrong from cell")
			assert.Equal(t, a1.ToCell, c.toCell, "wrong to cell")
		})
	}
}

func TestRawQueryRowsResult_toQueryRowsResult(t *testing.T) {
	t.Run("empty_rows", func(t *testing.T) {
		r := rawQueryRowsResult{
			Table: rawQueryRowsResultTable{
				Cols: []rawQueryRowsResultColumn{
					{ID: "A", Type: "number"},
					{ID: "B", Type: "string"},
				},
				Rows: []rawQueryRowsResultRow{},
			},
		}

		expected := QueryRowsResult{Rows: make([][]interface{}, 0)}

		result, err := r.toQueryRowsResult()
		assert.Nil(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("few_rows", func(t *testing.T) {
		r := rawQueryRowsResult{
			Table: rawQueryRowsResultTable{
				Cols: []rawQueryRowsResultColumn{
					{ID: "A", Type: "number"},
					{ID: "B", Type: "string"},
					{ID: "C", Type: "boolean"},
				},
				Rows: []rawQueryRowsResultRow{
					{
						[]rawQueryRowsResultCell{
							{Value: 123.0, Raw: "123"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "true"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 456.0, Raw: "456"},
							{Value: "blah2", Raw: "blah2"},
							{Value: false, Raw: "FALSE"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 123.1, Raw: "123.1"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "TRUE"},
						},
					},
				},
			},
		}

		expected := QueryRowsResult{
			Rows: [][]interface{}{
				{float64(123), "blah", true},
				{float64(456), "blah2", false},
				{123.1, "blah", true},
			},
		}

		result, err := r.toQueryRowsResult()
		assert.Nil(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("unexpected_type", func(t *testing.T) {
		r := rawQueryRowsResult{
			Table: rawQueryRowsResultTable{
				Cols: []rawQueryRowsResultColumn{
					{ID: "A", Type: "number"},
					{ID: "B", Type: "string"},
					{ID: "C", Type: "something"},
				},
				Rows: []rawQueryRowsResultRow{
					{
						[]rawQueryRowsResultCell{
							{Value: 123.0, Raw: "123"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "true"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 456.0, Raw: "456"},
							{Value: "blah2", Raw: "blah2"},
							{Value: false, Raw: "FALSE"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 123.1, Raw: "123.1"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "TRUE"},
						},
					},
				},
			},
		}

		result, err := r.toQueryRowsResult()
		assert.Equal(t, QueryRowsResult{}, result)
		assert.NotNil(t, err)
	})
}
