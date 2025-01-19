package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreeLeh/GoFreeDB/internal/models"
)

func TestBasicCodecEncode(t *testing.T) {
	tc := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "!",
		},
		{
			name:     "non_empty_string",
			input:    "test",
			expected: "!test",
		},
		{
			name:     "emoji",
			input:    "😀",
			expected: "!😀",
		},
		{
			name:     "NA_value",
			input:    models.NAValue,
			expected: "!" + models.NAValue,
		},
	}
	codec := &Basic{}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			result, err := codec.Encode([]byte(c.input))
			assert.Nil(t, err)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestBasicCodecDecode(t *testing.T) {
	tc := []struct {
		name     string
		input    string
		expected []byte
		hasErr   bool
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: []byte(nil),
			hasErr:   true,
		},
		{
			name:     "non_empty_string_no_whitespace",
			input:    "test",
			expected: []byte(nil),
			hasErr:   true,
		},
		{
			name:     "non_empty_string",
			input:    "!test",
			expected: []byte("test"),
			hasErr:   false,
		},
		{
			name:     "emoji",
			input:    "!😀",
			expected: []byte("😀"),
			hasErr:   false,
		},
		{
			name:     "NA_value",
			input:    "!" + models.NAValue,
			expected: []byte(models.NAValue),
			hasErr:   false,
		},
	}
	codec := &Basic{}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			result, err := codec.Decode(c.input)
			assert.Equal(t, c.hasErr, err != nil)
			assert.Equal(t, c.expected, result)
		})
	}
}
