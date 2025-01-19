package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEscapeValue(t *testing.T) {
	assert.Equal(t, "'blah", EscapeValue("blah"))
	assert.Equal(t, 1, EscapeValue(1))
	assert.Equal(t, true, EscapeValue(true))
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
