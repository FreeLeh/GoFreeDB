package freeleh

import (
	"context"
	"errors"
	"fmt"

	"github.com/FreeLeh/GoFreeLeh/internal/google/sheets"
)

// GoogleSheetKVStoreConfig defines a list of configurations that can be used to customise how the GoogleSheetKVStore works.
type GoogleSheetKVStoreConfig struct {
	Mode  KVMode
	codec Codec
}

// GoogleSheetKVStore encapsulates key-value store functionality on top of a Google Sheet.
//
// There are 2 operation modes for the key-value store: default and append only mode.
//
// For more details on how they differ, please read the explanations for each method or the TODO(edocsss) protocol page.
type GoogleSheetKVStore struct {
	wrapper             sheetsWrapper
	spreadsheetID       string
	sheetName           string
	scratchpadSheetName string
	scratchpadLocation  sheets.A1Range
	config              GoogleSheetKVStoreConfig
}

// Get retrieves the value associated with the given key.
// If the key exists in the store, the raw bytes value and no error will be returned.
// If the key does not exist in the store, a nil []byte and a wrapped ErrKeyNotFound will be returned.
//
// In default mode,
//     - There will be only one row with the given key. It will return the value for that in that row.
//     - There is only 1 API call behind the scene.
//
// In append only mode,
//     - As there could be multiple rows with the same key, we need to only use the latest row as it contains the last updated value.
//     - Note that deletion using append only mode results in a new row with a tombstone value. This method will also recognise and handle such cases.
//     - There is only 1 API call behind the scene.
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

// Set inserts the key-value pair into the key-value store.
//
// In default mode,
//     - If the key is not in the store, `Set` will create a new row and store the key value pair there.
//     - If the key is in the store, `Set` will update the previous row with the new value and timestamp.
//     - There are exactly 2 API calls behind the scene: getting the row for the key and creating/updating with the given key value data.
//
// In append only mode,
//     - It always creates a new row at the bottom of the sheet with the latest value and timestamp.
//     - There is only 1 API call behind the scene.
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

// Delete deletes the given key from the key-value store.
//
// In default mode,
//     - If the key is not in the store, it will not do anything.
//     - If the key is in the store, it will remove that row.
//     - There are up to 2 API calls behind the scene: getting the row for the key and remove the row (if the key exists).
//
// In append only mode,
//     - It creates a new row at the bottom of the sheet with a tombstone value and timestamp.
//     - There is only 1 API call behind the scene.
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

// Close cleans up all held resources like the scratchpad cell booked for this specific GoogleSheetKVStore instance.
func (s *GoogleSheetKVStore) Close(ctx context.Context) error {
	_, err := s.wrapper.Clear(ctx, s.spreadsheetID, []string{s.scratchpadLocation.Original})
	return err
}

// NewGoogleSheetKVStore creates an instance of the key-value store with the given configuration.
// It will also try to create the sheet, in case it does not exist yet.
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
	config.codec = &basicCodec{}
	return config
}
