package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetA1Range(t *testing.T) {
	assert.Equal(t, "sheet!A1:A50", NewA1Range("sheet", "A1:A50").Original)
	assert.Equal(t, "sheet!A1", NewA1Range("sheet", "A1").Original)
	assert.Equal(t, "sheet!A", NewA1Range("sheet", "A").Original)
}

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
			a1 := NewA1RangeFromString(c.input)
			assert.Equal(t, a1.Original, c.input, "A1Range original should have the same value as the input")
			assert.Equal(t, a1.SheetName, c.sheetName, "wrong sheet name")
			assert.Equal(t, a1.FromCell, c.fromCell, "wrong from cell")
			assert.Equal(t, a1.ToCell, c.toCell, "wrong to cell")
		})
	}
}

func TestGenerateColumnName(t *testing.T) {
	tc := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "A",
		},
		{
			name:     "single_character",
			input:    15,
			expected: "P",
		},
		{
			name:     "single_character_2",
			input:    25,
			expected: "Z",
		},
		{
			name:     "single_character_3",
			input:    5,
			expected: "F",
		},
		{
			name:     "double_character",
			input:    26,
			expected: "AA",
		},
		{
			name:     "double_character_2",
			input:    52,
			expected: "BA",
		},
		{
			name:     "double_character_2",
			input:    89,
			expected: "CL",
		},
		{
			name:     "max_column",
			input:    18277,
			expected: "ZZZ",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, GenerateColumnName(c.input))
		})
	}
}

func TestCellToColIdx(t *testing.T) {
	tc := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "single_char_A",
			input:    "A",
			expected: 1,
		},
		{
			name:     "single_char_Z",
			input:    "Z",
			expected: 26,
		},
		{
			name:     "single_char_with_number",
			input:    "A1",
			expected: 1,
		},
		{
			name:     "double_char_AA",
			input:    "AA",
			expected: 27,
		},
		{
			name:     "double_char_AZ",
			input:    "AZ",
			expected: 52,
		},
		{
			name:     "double_char_BA",
			input:    "BA",
			expected: 53,
		},
		{
			name:     "triple_char_AAA",
			input:    "AAA",
			expected: 703,
		},
		{
			name:     "triple_char_ABC",
			input:    "ABC",
			expected: 731,
		},
		{
			name:     "lowercase_input",
			input:    "abc",
			expected: 731,
		},
		{
			name:     "mixed_case_input",
			input:    "aBc",
			expected: 731,
		},
		{
			name:     "empty_string",
			input:    "",
			expected: 0,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, CellToColIdx(c.input))
		})
	}
}

func TestA1Range_NumCols(t *testing.T) {
	tc := []struct {
		name     string
		fromCell string
		toCell   string
		expected int
	}{
		{
			name:     "single_cell",
			fromCell: "A",
			toCell:   "A",
			expected: 1,
		},
		{
			name:     "single_column_range",
			fromCell: "A1",
			toCell:   "A10",
			expected: 1,
		},
		{
			name:     "single_char_range",
			fromCell: "A",
			toCell:   "C",
			expected: 3,
		},
		{
			name:     "multi_char_range",
			fromCell: "AA",
			toCell:   "AC",
			expected: 3,
		},
		{
			name:     "large_range",
			fromCell: "A",
			toCell:   "Z",
			expected: 26,
		},
		{
			name:     "double_char_to_triple_char",
			fromCell: "AZ",
			toCell:   "BA",
			expected: 2,
		},
		{
			name:     "triple_char_range",
			fromCell: "AAA",
			toCell:   "AAC",
			expected: 3,
		},
		{
			name:     "reverse_range",
			fromCell: "C",
			toCell:   "A",
			expected: 3,
		},
		{
			name:     "same_cell_with_numbers",
			fromCell: "B5",
			toCell:   "B5",
			expected: 1,
		},
		{
			name:     "different_cells_with_numbers",
			fromCell: "B5",
			toCell:   "D5",
			expected: 3,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			r := A1Range{
				FromCell: c.fromCell,
				ToCell:   c.toCell,
			}
			assert.Equal(t, c.expected, r.NumCols())
		})
	}
}

func TestA1Range_NumRows(t *testing.T) {
	tc := []struct {
		name     string
		fromCell string
		toCell   string
		expected int
	}{
		{
			name:     "single_cell",
			fromCell: "A1",
			toCell:   "A1",
			expected: 1,
		},
		{
			name:     "single_column_range",
			fromCell: "A1",
			toCell:   "A10",
			expected: 10,
		},
		{
			name:     "multi_column_range",
			fromCell: "A1",
			toCell:   "C10",
			expected: 10,
		},
		{
			name:     "reverse_range",
			fromCell: "A10",
			toCell:   "A1",
			expected: 10,
		},
		{
			name:     "large_range",
			fromCell: "A100",
			toCell:   "A200",
			expected: 101,
		},
		{
			name:     "no_numbers_in_cells",
			fromCell: "A",
			toCell:   "C",
			expected: 0,
		},
		{
			name:     "invalid_number_in_from_cell",
			fromCell: "Aabc",
			toCell:   "A10",
			expected: 0,
		},
		{
			name:     "invalid_number_in_to_cell",
			fromCell: "A1",
			toCell:   "Axyz",
			expected: 0,
		},
		{
			name:     "empty_cells",
			fromCell: "",
			toCell:   "",
			expected: 0,
		},
		{
			name:     "from_cell_without_number",
			fromCell: "A",
			toCell:   "A10",
			expected: 0,
		},
		{
			name:     "to_cell_without_number",
			fromCell: "A1",
			toCell:   "A",
			expected: 0,
		},
		{
			name:     "invalid_chars_after_digit",
			fromCell: "A1ajskhd",
			toCell:   "A100",
			expected: 0,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			r := A1Range{
				FromCell: c.fromCell,
				ToCell:   c.toCell,
			}
			assert.Equal(t, c.expected, r.NumRows())
		})
	}
}

func TestGenerateColumnMapping(t *testing.T) {
	tc := []struct {
		name     string
		input    []string
		expected map[string]ColIdx
	}{
		{
			name:  "single_column",
			input: []string{"col1"},
			expected: map[string]ColIdx{
				"col1": {"A", 0},
			},
		},
		{
			name:  "three_column",
			input: []string{"col1", "col2", "col3"},
			expected: map[string]ColIdx{
				"col1": {"A", 0},
				"col2": {"B", 1},
				"col3": {"C", 2},
			},
		},
		{
			name: "many_column",
			input: []string{
				"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10",
				"c11", "c12", "c13", "c14", "c15", "c16", "c17", "c18", "c19", "c20",
				"c21", "c22", "c23", "c24", "c25", "c26", "c27", "c28",
			},
			expected: map[string]ColIdx{
				"c1": {"A", 0}, "c2": {"B", 1}, "c3": {"C", 2}, "c4": {"D", 3},
				"c5": {"E", 4}, "c6": {"F", 5}, "c7": {"G", 6}, "c8": {"H", 7},
				"c9": {"I", 8}, "c10": {"J", 9}, "c11": {"K", 10}, "c12": {"L", 11},
				"c13": {"M", 12}, "c14": {"N", 13}, "c15": {"O", 14}, "c16": {"P", 15},
				"c17": {"Q", 16}, "c18": {"R", 17}, "c19": {"S", 18}, "c20": {"T", 19},
				"c21": {"U", 20}, "c22": {"V", 21}, "c23": {"W", 22}, "c24": {"X", 23},
				"c25": {"Y", 24}, "c26": {"Z", 25}, "c27": {"AA", 26}, "c28": {"AB", 27},
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, GenerateColumnMapping(c.input))
		})
	}
}
