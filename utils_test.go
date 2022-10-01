package freedb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetA1Range(t *testing.T) {
	assert.Equal(t, "sheet!A1:A50", getA1Range("sheet", "A1:A50"))
	assert.Equal(t, "sheet!A1", getA1Range("sheet", "A1"))
	assert.Equal(t, "sheet!A", getA1Range("sheet", "A"))
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
			assert.Equal(t, c.expected, generateColumnName(c.input))
		})
	}
}

func TestGenerateColumnMapping(t *testing.T) {
	tc := []struct {
		name     string
		input    []string
		expected map[string]colIdx
	}{
		{
			name:  "single_column",
			input: []string{"col1"},
			expected: map[string]colIdx{
				"col1": {"A", 0},
			},
		},
		{
			name:  "three_column",
			input: []string{"col1", "col2", "col3"},
			expected: map[string]colIdx{
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
			expected: map[string]colIdx{
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
			assert.Equal(t, c.expected, generateColumnMapping(c.input))
		})
	}
}

func TestEscapeValue(t *testing.T) {
	assert.Equal(t, "'blah", escapeValue("blah"))
	assert.Equal(t, 1, escapeValue(1))
	assert.Equal(t, true, escapeValue(true))
}

func TestCheckIEEE754SafeInteger(t *testing.T) {
	assert.Nil(t, checkIEEE754SafeInteger(int64(0)))
	assert.Nil(t, checkIEEE754SafeInteger(int(0)))
	assert.Nil(t, checkIEEE754SafeInteger(uint(0)))

	// -(2^53)
	assert.Nil(t, checkIEEE754SafeInteger(int64(-9007199254740992)))

	// (2^53)
	assert.Nil(t, checkIEEE754SafeInteger(int64(9007199254740992)))
	assert.Nil(t, checkIEEE754SafeInteger(uint64(9007199254740992)))

	// Below and above the limit.
	assert.NotNil(t, checkIEEE754SafeInteger(int64(-9007199254740993)))
	assert.NotNil(t, checkIEEE754SafeInteger(int64(9007199254740993)))
	assert.NotNil(t, checkIEEE754SafeInteger(uint64(9007199254740993)))

	// Other types
	assert.Nil(t, checkIEEE754SafeInteger("blah"))
	assert.Nil(t, checkIEEE754SafeInteger(true))
	assert.Nil(t, checkIEEE754SafeInteger([]byte("something")))
}
