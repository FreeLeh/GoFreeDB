package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEscapeValue(t *testing.T) {
	t.Run("not in cols with formula", func(t *testing.T) {
		value, err := EscapeValue("A", 123, NewSet([]string{"B"}))
		assert.Nil(t, err)
		assert.Equal(t, 123, value)

		value, err = EscapeValue("A", "123", NewSet([]string{"B"}))
		assert.Nil(t, err)
		assert.Equal(t, "'123", value)
	})

	t.Run("in cols with formula, but not string", func(t *testing.T) {
		value, err := EscapeValue("A", 123, NewSet([]string{"A"}))
		assert.NotNil(t, err)
		assert.Equal(t, nil, value)
	})

	t.Run("in cols with formula, but string", func(t *testing.T) {
		value, err := EscapeValue("A", "123", NewSet([]string{"A"}))
		assert.Nil(t, err)
		assert.Equal(t, "123", value)
	})

	t.Run("different data types, not in the cols with formula", func(t *testing.T) {
		value, err := EscapeValue("A", "blah", NewSet([]string{}))
		assert.Equal(t, "'blah", value)
		assert.NoError(t, err)

		value, err = EscapeValue("A", 1, NewSet([]string{}))
		assert.Equal(t, 1, value)
		assert.NoError(t, err)

		value, err = EscapeValue("A", true, NewSet([]string{}))
		assert.Equal(t, true, value)
		assert.NoError(t, err)
	})
}

func TestCheckIEEE754SafeInteger(t *testing.T) {
	assert.Nil(t, CheckIEEE754SafeInteger(int64(0)))
	assert.Nil(t, CheckIEEE754SafeInteger(int(0)))
	assert.Nil(t, CheckIEEE754SafeInteger(uint(0)))

	// -(2^53)
	assert.Nil(t, CheckIEEE754SafeInteger(int64(-9007199254740992)))

	// (2^53)
	assert.Nil(t, CheckIEEE754SafeInteger(int64(9007199254740992)))
	assert.Nil(t, CheckIEEE754SafeInteger(uint64(9007199254740992)))

	// Below and above the limit.
	assert.NotNil(t, CheckIEEE754SafeInteger(int64(-9007199254740993)))
	assert.NotNil(t, CheckIEEE754SafeInteger(int64(9007199254740993)))
	assert.NotNil(t, CheckIEEE754SafeInteger(uint64(9007199254740993)))

	// Other types
	assert.Nil(t, CheckIEEE754SafeInteger("blah"))
	assert.Nil(t, CheckIEEE754SafeInteger(true))
	assert.Nil(t, CheckIEEE754SafeInteger([]byte("something")))
}
