package freeleh

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FreeLeh/GoFreeLeh/internal/google/sheets"
)

type GoogleSheetKVConfig struct {
	Mode  KVMode
	codec Codec
}

type Codec interface {
	Encode(value []byte) (string, error)
	Decode(value string) ([]byte, error)
}

type Sheets interface {
	CreateSpreadsheet(ctx context.Context, title string) (string, error)
	CreateSheet(ctx context.Context, spreadsheetID string, sheetName string) error
	InsertRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (sheets.InsertRowsResult, error)
	OverwriteRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (sheets.InsertRowsResult, error)
	UpdateRows(ctx context.Context, spreadsheetID string, a1Range string, values [][]interface{}) (sheets.UpdateRowsResult, error)
	Clear(ctx context.Context, spreadsheetID string, ranges []string) ([]string, error)
}

/*

There are 2 formats of the same KV storage using Google Sheet:
- Default -> like a normal KV store, each key only appears once in the sheet.
- Append only -> each key update will be added as a new row, there maybe >1 rows for the same keys. The latest added row for a key is the latest value.

## APPEND ONLY MODE

The structure is as follows:

key	| value | timestamp
k1	| v1	| 1
k2	| v2	| 2
k3	| v3	| 3
k2	| v22	| 4 			--> Set(k2, v22)
k3	| v32	| 5				--> Set(k3, v32)
k2	| 		| 6				--> Delete(k2) -> value is set to an empty string

The logic for Set() is as simple as appending a new row at the end of the current sheet with the latest value and the timestamp in milliseconds.
The logic for Delete() is basically Set(key, "").
The logic for Get() is more complicated.

=VLOOKUP(key, SORT(<full_table_range>, 3, FALSE), 2, FALSE)

The full table range can be assumed to be A1:C5000000.
The integer "3" is referring to the timestamp column (the third column of the table).
The FALSE inside the SORT() means sort in descending order.
The FALSE inside the VLOOKUP() means we consider the table as non-sorted (so that we can take the first row which has the latest timestamp as the final value).
The integer "2" is referring to which column we want to return from VLOOKUP(), which is referring to the value column.

If the value returned is either "#N/A" or "", that means the key is not found or already deleted.

## DEFAULT MODE

The structure is as follows:

key	| value | timestamp
k1	| v1	| 1
k2	| v2	| 2
k3	| v3	| 3

The logic for Set() is Get() + (Append(OVERWRITE_MODE) if not exists OR Update if already exists).
The logic for Delete() is Get() + Clear().
The logic for Get() is just a simple VLOOKUP without any sorting involved (unlike the APPEND ONLY mode).

Here we assume there cannot be any race condition that leads to two rows with the same key.

*/
type GoogleSheetKV struct {
	wrapper             Sheets
	spreadsheetID       string
	sheetName           string
	scratchpadSheetName string
	scratchpadLocation  sheets.A1Range
	config              GoogleSheetKVConfig
}

func (kv *GoogleSheetKV) Get(ctx context.Context, key string) ([]byte, error) {
	query := fmt.Sprintf(getDefaultQueryTemplate, key, getA1Range(kv.sheetName, defaultTableRange))
	if kv.config.Mode == KVModeAppendOnly {
		query = fmt.Sprintf(getAppendQueryTemplate, key, getA1Range(kv.sheetName, defaultTableRange))
	}

	result, err := kv.wrapper.UpdateRows(
		ctx,
		kv.spreadsheetID,
		kv.scratchpadLocation.Original,
		[][]interface{}{{query}},
	)
	if err != nil {
		return nil, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	value := result.UpdatedValues[0][0]
	if value == naValue || value == "" {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	return kv.config.codec.Decode(value.(string))
}

func (kv *GoogleSheetKV) Set(ctx context.Context, key string, value []byte) error {
	encoded, err := kv.config.codec.Encode(value)
	if err != nil {
		return err
	}
	if kv.config.Mode == KVModeAppendOnly {
		return kv.setAppendOnly(ctx, key, encoded)
	}
	return kv.setDefault(ctx, key, encoded)
}

func (kv *GoogleSheetKV) setAppendOnly(ctx context.Context, key string, encoded string) error {
	_, err := kv.wrapper.InsertRows(
		ctx,
		kv.spreadsheetID,
		getA1Range(kv.sheetName, defaultTableRange),
		[][]interface{}{{key, encoded, currentTimeMs()}},
	)
	return err
}

func (kv *GoogleSheetKV) setDefault(ctx context.Context, key string, encoded string) error {
	a1Range, err := kv.findKeyA1Range(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		_, err := kv.wrapper.OverwriteRows(
			ctx,
			kv.spreadsheetID,
			getA1Range(kv.sheetName, defaultTableRange),
			[][]interface{}{{key, encoded, currentTimeMs()}},
		)
		return err
	}

	if err != nil {
		return err
	}

	_, err = kv.wrapper.UpdateRows(
		ctx,
		kv.spreadsheetID,
		a1Range.Original,
		[][]interface{}{{key, encoded, currentTimeMs()}},
	)
	return err
}

func (kv *GoogleSheetKV) findKeyA1Range(ctx context.Context, key string) (sheets.A1Range, error) {
	result, err := kv.wrapper.UpdateRows(
		ctx,
		kv.spreadsheetID,
		kv.scratchpadLocation.Original,
		[][]interface{}{{fmt.Sprintf(findKeyA1RangeQueryTemplate, key, getA1Range(kv.sheetName, defaultKeyColRange))}},
	)
	if err != nil {
		return sheets.A1Range{}, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return sheets.A1Range{}, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	offset := result.UpdatedValues[0][0]
	if offset == naValue || offset == "" {
		return sheets.A1Range{}, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	// Note that the MATCH() query only returns the relative offset from the given range.
	// Here we need to return the full range where the key is found.
	// Hence, we need to get the row offset first, and assume that each row has only 3 rows: A B C.
	// Otherwise, the DELETE() function will not work properly (we need to clear the full row, not just the key cell).
	a1Range := getA1Range(kv.sheetName, fmt.Sprintf("A%s:C%s", offset, offset))
	return sheets.NewA1Range(a1Range), nil
}

func (kv *GoogleSheetKV) Delete(ctx context.Context, key string) error {
	if kv.config.Mode == KVModeAppendOnly {
		return kv.deleteAppendOnly(ctx, key)
	}
	return kv.deleteDefault(ctx, key)
}

func (kv *GoogleSheetKV) deleteAppendOnly(ctx context.Context, key string) error {
	return kv.setAppendOnly(ctx, key, "")
}

func (kv *GoogleSheetKV) deleteDefault(ctx context.Context, key string) error {
	a1Range, err := kv.findKeyA1Range(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = kv.wrapper.Clear(ctx, kv.spreadsheetID, []string{a1Range.Original})
	return err
}

func (kv *GoogleSheetKV) findScratchpadLocation() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	result, err := kv.wrapper.OverwriteRows(
		ctx,
		kv.spreadsheetID,
		kv.scratchpadSheetName+"!"+defaultTableRange,
		[][]interface{}{{scratchpadBooked}},
	)
	if err != nil {
		return err
	}

	kv.scratchpadLocation = result.UpdatedRange
	return nil
}

func (kv *GoogleSheetKV) ensureSheets() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	kv.wrapper.CreateSheet(ctx, kv.spreadsheetID, kv.sheetName)
	kv.wrapper.CreateSheet(ctx, kv.spreadsheetID, kv.scratchpadSheetName)
}

func (kv *GoogleSheetKV) Close(ctx context.Context) error {
	_, err := kv.wrapper.Clear(ctx, kv.spreadsheetID, []string{kv.scratchpadLocation.Original})
	return err
}

func NewGoogleSheetKeyValue(
	auth sheets.AuthClient,
	spreadsheetID string,
	sheetName string,
	config GoogleSheetKVConfig,
) *GoogleSheetKV {
	wrapper, err := sheets.NewWrapper(auth)
	if err != nil {
		panic(err)
	}

	scratchpadSheetName := sheetName + scratchpadSheetNameSuffix
	config = applyConfig(config)

	kv := &GoogleSheetKV{
		wrapper:             wrapper,
		spreadsheetID:       spreadsheetID,
		sheetName:           sheetName,
		scratchpadSheetName: scratchpadSheetName,
		config:              config,
	}

	kv.ensureSheets()
	if err := kv.findScratchpadLocation(); err != nil {
		panic(err)
	}
	return kv
}

func applyConfig(config GoogleSheetKVConfig) GoogleSheetKVConfig {
	config.codec = &BasicCodec{}
	return config
}
