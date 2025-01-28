package common

import (
	"errors"
	"fmt"
)

func EscapeValue(
	col string,
	value any,
	colsWithFormula *Set[string],
) (any, error) {
	if !colsWithFormula.Contains(col) {
		// This is to ensure that string value will always be a string representation in Google Sheets.
		// Without this, "1" may be converted automatically into an integer.
		// "2020-01-01" may be converted into a date format.
		switch value.(type) {
		case string:
			return fmt.Sprintf("'%s", value), nil
		default:
			return value, nil
		}
	}

	_, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("value of column %s is not a string, but expected to contain formula", col)
	}
	return value, nil
}

func CheckIEEE754SafeInteger(value interface{}) error {
	switch converted := value.(type) {
	case int:
		return isIEEE754SafeInteger(int64(converted))
	case int64:
		return isIEEE754SafeInteger(converted)
	case uint:
		return isIEEE754SafeInteger(int64(converted))
	case uint64:
		return isIEEE754SafeInteger(int64(converted))
	default:
		return nil
	}
}

func isIEEE754SafeInteger(value int64) error {
	if value == int64(float64(value)) {
		return nil
	}
	return errors.New("integer provided is not within the IEEE 754 safe integer boundary of [-(2^53), 2^53], the integer may have a precision lost")
}
