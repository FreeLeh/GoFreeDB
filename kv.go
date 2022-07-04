package freeleh

import (
	"context"
	"errors"
	"fmt"

	"github.com/FreeLeh/GoFreeLeh/internal/google/sheets"
)

type GoogleSheetKVStoreConfig struct {
	Mode  KVMode
	codec Codec
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
type GoogleSheetKVStore struct {
	wrapper             sheetsWrapper
	spreadsheetID       string
	sheetName           string
	scratchpadSheetName string
	scratchpadLocation  sheets.A1Range
	config              GoogleSheetKVStoreConfig
}

func (s *GoogleSheetKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	query := fmt.Sprintf(kvGetDefaultQueryTemplate, key, getA1Range(s.sheetName, defaultKVTableRange))
	if s.config.Mode == KVModeAppendOnly {
		query = fmt.Sprintf(kvGetAppendQueryTemplate, key, getA1Range(s.sheetName, defaultKVTableRange))
	}

	result, err := s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		s.scratchpadLocation.Original,
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
	return s.config.codec.Decode(value.(string))
}

func (s *GoogleSheetKVStore) Set(ctx context.Context, key string, value []byte) error {
	encoded, err := s.config.codec.Encode(value)
	if err != nil {
		return err
	}
	if s.config.Mode == KVModeAppendOnly {
		return s.setAppendOnly(ctx, key, encoded)
	}
	return s.setDefault(ctx, key, encoded)
}

func (s *GoogleSheetKVStore) setAppendOnly(ctx context.Context, key string, encoded string) error {
	_, err := s.wrapper.InsertRows(
		ctx,
		s.spreadsheetID,
		getA1Range(s.sheetName, defaultKVTableRange),
		[][]interface{}{{key, encoded, currentTimeMs()}},
	)
	return err
}

func (s *GoogleSheetKVStore) setDefault(ctx context.Context, key string, encoded string) error {
	a1Range, err := s.findKeyA1Range(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		_, err := s.wrapper.OverwriteRows(
			ctx,
			s.spreadsheetID,
			getA1Range(s.sheetName, defaultKVFirstRowRange),
			[][]interface{}{{key, encoded, currentTimeMs()}},
		)
		return err
	}

	if err != nil {
		return err
	}

	_, err = s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		a1Range.Original,
		[][]interface{}{{key, encoded, currentTimeMs()}},
	)
	return err
}

func (s *GoogleSheetKVStore) findKeyA1Range(ctx context.Context, key string) (sheets.A1Range, error) {
	result, err := s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		s.scratchpadLocation.Original,
		[][]interface{}{{fmt.Sprintf(kvFindKeyA1RangeQueryTemplate, key, getA1Range(s.sheetName, defaultKVKeyColRange))}},
	)
	if err != nil {
		return sheets.A1Range{}, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return sheets.A1Range{}, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	offset := result.UpdatedValues[0][0].(string)
	if offset == naValue || offset == "" {
		return sheets.A1Range{}, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	// Note that the MATCH() query only returns the relative offset from the given range.
	// Here we need to return the full range where the key is found.
	// Hence, we need to get the row offset first, and assume that each row has only 3 rows: A B C.
	// Otherwise, the DELETE() function will not work properly (we need to clear the full row, not just the key cell).
	a1Range := getA1Range(s.sheetName, fmt.Sprintf("A%s:C%s", offset, offset))
	return sheets.NewA1Range(a1Range), nil
}

func (s *GoogleSheetKVStore) Delete(ctx context.Context, key string) error {
	if s.config.Mode == KVModeAppendOnly {
		return s.deleteAppendOnly(ctx, key)
	}
	return s.deleteDefault(ctx, key)
}

func (s *GoogleSheetKVStore) deleteAppendOnly(ctx context.Context, key string) error {
	return s.setAppendOnly(ctx, key, "")
}

func (s *GoogleSheetKVStore) deleteDefault(ctx context.Context, key string) error {
	a1Range, err := s.findKeyA1Range(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = s.wrapper.Clear(ctx, s.spreadsheetID, []string{a1Range.Original})
	return err
}

func (s *GoogleSheetKVStore) Close(ctx context.Context) error {
	_, err := s.wrapper.Clear(ctx, s.spreadsheetID, []string{s.scratchpadLocation.Original})
	return err
}

func NewGoogleSheetKVStore(
	auth sheets.AuthClient,
	spreadsheetID string,
	sheetName string,
	config GoogleSheetKVStoreConfig,
) *GoogleSheetKVStore {
	wrapper, err := sheets.NewWrapper(auth)
	if err != nil {
		panic(fmt.Errorf("error creating sheets wrapper: %w", err))
	}

	scratchpadSheetName := sheetName + scratchpadSheetNameSuffix
	config = applyGoogleSheetKVStoreConfig(config)

	store := &GoogleSheetKVStore{
		wrapper:             wrapper,
		spreadsheetID:       spreadsheetID,
		sheetName:           sheetName,
		scratchpadSheetName: scratchpadSheetName,
		config:              config,
	}

	_ = ensureSheets(store.wrapper, store.spreadsheetID, store.sheetName)
	_ = ensureSheets(store.wrapper, store.spreadsheetID, store.scratchpadSheetName)

	scratchpadLocation, err := findScratchpadLocation(store.wrapper, store.spreadsheetID, store.scratchpadSheetName)
	if err != nil {
		panic(fmt.Errorf("error finding a scratchpad location in sheet %s: %w", store.scratchpadSheetName, err))
	}
	store.scratchpadLocation = scratchpadLocation

	return store
}

func applyGoogleSheetKVStoreConfig(config GoogleSheetKVStoreConfig) GoogleSheetKVStoreConfig {
	config.codec = &BasicCodec{}
	return config
}
