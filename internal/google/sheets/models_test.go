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
