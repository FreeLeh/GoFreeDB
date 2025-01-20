package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONEncodeNoError(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "simple string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "integer",
			input:    42,
			expected: `42`,
		},
		{
			name:     "float",
			input:    3.14,
			expected: `3.14`,
		},
		{
			name:     "boolean",
			input:    true,
			expected: `true`,
		},
		{
			name:     "struct",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			expected: `{"name":"John","age":30}`,
		},
		{
			name:     "slice",
			input:    []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
		{
			name:     "map",
			input:    map[string]int{"a": 1, "b": 2},
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "time",
			input:    time.Date(2025, 1, 20, 15, 30, 0, 0, time.UTC),
			expected: `"2025-01-20T15:30:00Z"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONEncodeNoError(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONEncodeNoError_InvalidInput(t *testing.T) {
	// Channel cannot be JSON encoded
	ch := make(chan int)
	result := JSONEncodeNoError(ch)
	assert.Equal(t, "", result)
}
