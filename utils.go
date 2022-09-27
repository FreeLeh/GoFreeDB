package freedb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
	"github.com/mitchellh/mapstructure"
)

func currentTimeMs() int64 {
	return time.Now().UnixMilli()
}

func getA1Range(sheetName string, rng string) string {
	return sheetName + "!" + rng
}

func ensureSheets(wrapper sheetsWrapper, spreadsheetID string, sheetName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	return wrapper.CreateSheet(ctx, spreadsheetID, sheetName)
}

func findScratchpadLocation(wrapper sheetsWrapper, spreadsheetID string, scratchpadSheetName string) (sheets.A1Range, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	result, err := wrapper.OverwriteRows(
		ctx,
		spreadsheetID,
		scratchpadSheetName+"!"+defaultKVTableRange,
		[][]interface{}{{scratchpadBooked}},
	)
	if err != nil {
		return sheets.A1Range{}, err
	}
	return result.UpdatedRange, nil
}

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateColumnMapping(columns []string) map[string]colIdx {
	mapping := make(map[string]colIdx, len(columns))
	for n, col := range columns {
		mapping[col] = colIdx{
			name: generateColumnName(n),
			idx:  n,
		}
	}
	return mapping
}

// This is not purely a Base26 conversion since the second char can start from "A" (or 0) again.
// In a normal Base26 int to string conversion, the second char can only start from "B" (or 1).
// Hence, we need to hack it by checking the first round separately from the subsequent round.
// For the subsequent rounds, we need to subtract by 1 first or else it will always start from 1 (not 0).
func generateColumnName(n int) string {
	col := string(alphabet[n%26])
	n = n / 26

	for {
		if n <= 0 {
			break
		}

		n -= 1
		col = string(alphabet[n%26]) + col
		n = n / 26
	}

	return col
}

func mapstructureDecode(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Result:  output,
		TagName: "db",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// This is to ensure that string value will always be a string representation in Google Sheets.
// Without this, "1" may be converted automatically into an integer.
// "2020-01-01" may be converted into a date format.
func escapeValue(value interface{}) interface{} {
	switch value.(type) {
	case string:
		return fmt.Sprintf("'%s", value)
	default:
		return value
	}
}

func checkIEEE754SafeInteger(value interface{}) error {
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
