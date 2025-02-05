package google

import (
	"context"
	"errors"
	"fmt"

	"github.com/FreeLeh/GoFreeDB/internal/codec"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

// SheetKVStoreConfig defines a list of configurations that can be used to customise how the SheetKVStore works.
type SheetKVStoreConfig struct {
	Mode  models.KVMode
	codec Codec
}

// SheetKVStore encapsulates key-value store functionality on top of a Google Sheet.
//
// There are 2 operation modes for the key-value store: default and append only mode.
//
// For more details on how they differ, please read the explanations for each method or the protocol page:
// https://github.com/FreeLeh/docs/blob/main/freedb/protocols.md.
type SheetKVStore struct {
	wrapper             sheetsWrapper
	spreadsheetID       string
	sheetName           string
	scratchpadSheetName string
	scratchpadLocation  models.A1Range
	config              SheetKVStoreConfig
}

// Get retrieves the value associated with the given key.
// If the key exists in the store, the raw bytes value and no error will be returned.
// If the key does not exist in the store, a nil []byte and a wrapped ErrKeyNotFound will be returned.
//
// In default mode,
//   - There will be only one row with the given key. It will return the value for that in that row.
//   - There is only 1 API call behind the scene.
//
// In append only mode,
//   - As there could be multiple rows with the same key, we need to only use the latest row as it
//     contains the last updated value.
//   - Note that deletion using append only mode results in a new row with a tombstone value.
//     This method will also recognise and handle such cases.
//   - There is only 1 API call behind the scene.
func (s *SheetKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	query := fmt.Sprintf(
		kvGetDefaultQueryTemplate,
		key,
		models.NewA1Range(s.sheetName, defaultKVTableRange).Original,
	)
	if s.config.Mode == models.KVModeAppendOnly {
		query = fmt.Sprintf(
			kvGetAppendQueryTemplate,
			key,
			models.NewA1Range(s.sheetName, defaultKVTableRange).Original,
		)
	}

	result, err := s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		s.scratchpadLocation,
		[][]interface{}{{query}},
	)
	if err != nil {
		return nil, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return nil, fmt.Errorf("%w: %s", models.ErrKeyNotFound, key)
	}

	value := result.UpdatedValues[0][0]
	if value == models.NAValue || value == "" {
		return nil, fmt.Errorf("%w: %s", models.ErrKeyNotFound, key)
	}
	return s.config.codec.Decode(value.(string))
}

// Set inserts the key-value pair into the key-value
//
// In default mode,
//   - If the key is not in the store, `Set` will create a new row and store the key value pair there.
//   - If the key is in the store, `Set` will update the previous row with the new value and timestamp.
//   - There are exactly 2 API calls behind the scene: getting the row for the key and creating/updating with the given key value data.
//
// In append only mode,
//   - It always creates a new row at the bottom of the sheet with the latest value and timestamp.
//   - There is only 1 API call behind the scene.
func (s *SheetKVStore) Set(ctx context.Context, key string, value []byte) error {
	encoded, err := s.config.codec.Encode(value)
	if err != nil {
		return err
	}
	if s.config.Mode == models.KVModeAppendOnly {
		return s.setAppendOnly(ctx, key, encoded)
	}
	return s.setDefault(ctx, key, encoded)
}

func (s *SheetKVStore) setAppendOnly(ctx context.Context, key string, encoded string) error {
	_, err := s.wrapper.InsertRows(
		ctx,
		s.spreadsheetID,
		models.NewA1Range(s.sheetName, defaultKVTableRange),
		[][]interface{}{{key, encoded, common.CurrentTimeMs()}},
	)
	return err
}

func (s *SheetKVStore) setDefault(ctx context.Context, key string, encoded string) error {
	a1Range, err := s.findKeyA1Range(ctx, key)
	if errors.Is(err, models.ErrKeyNotFound) {
		_, err := s.wrapper.OverwriteRows(
			ctx,
			s.spreadsheetID,
			models.NewA1Range(s.sheetName, defaultKVFirstRowRange),
			[][]interface{}{{key, encoded, common.CurrentTimeMs()}},
		)
		return err
	}

	if err != nil {
		return err
	}

	_, err = s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		a1Range,
		[][]interface{}{{key, encoded, common.CurrentTimeMs()}},
	)
	return err
}

func (s *SheetKVStore) findKeyA1Range(ctx context.Context, key string) (models.A1Range, error) {
	result, err := s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		s.scratchpadLocation,
		[][]interface{}{{fmt.Sprintf(
			kvFindKeyA1RangeQueryTemplate,
			key,
			models.NewA1Range(s.sheetName, defaultKVKeyColRange).Original,
		)}},
	)
	if err != nil {
		return models.A1Range{}, err
	}
	if len(result.UpdatedValues) == 0 || len(result.UpdatedValues[0]) == 0 {
		return models.A1Range{}, fmt.Errorf("%w: %s", models.ErrKeyNotFound, key)
	}

	offset := result.UpdatedValues[0][0].(string)
	if offset == models.NAValue || offset == "" {
		return models.A1Range{}, fmt.Errorf("%w: %s", models.ErrKeyNotFound, key)
	}

	// Note that the MATCH() query only returns the relative offset from the given range.
	// Here we need to return the full range where the key is found.
	// Hence, we need to get the row offset first, and assume that each row has only 3 rows: A B C.
	// Otherwise, the DELETE() function will not work properly (we need to clear the full row, not just the key cell).
	a1Range := models.NewA1Range(s.sheetName, fmt.Sprintf("A%s:C%s", offset, offset))
	return a1Range, nil
}

// Delete deletes the given key from the key-value
//
// In default mode,
//   - If the key is not in the store, it will not do anything.
//   - If the key is in the store, it will remove that row.
//   - There are up to 2 API calls behind the scene: getting the row for the key and remove the row (if the key exists).
//
// In append only mode,
//   - It creates a new row at the bottom of the sheet with a tombstone value and timestamp.
//   - There is only 1 API call behind the scene.
func (s *SheetKVStore) Delete(ctx context.Context, key string) error {
	if s.config.Mode == models.KVModeAppendOnly {
		return s.deleteAppendOnly(ctx, key)
	}
	return s.deleteDefault(ctx, key)
}

func (s *SheetKVStore) deleteAppendOnly(ctx context.Context, key string) error {
	return s.setAppendOnly(ctx, key, "")
}

func (s *SheetKVStore) deleteDefault(ctx context.Context, key string) error {
	a1Range, err := s.findKeyA1Range(ctx, key)
	if errors.Is(err, models.ErrKeyNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = s.wrapper.Clear(ctx, s.spreadsheetID, []models.A1Range{a1Range})
	return err
}

// Close cleans up all held resources like the scratchpad cell booked for this specific SheetKVStore instance.
func (s *SheetKVStore) Close(ctx context.Context) error {
	_, err := s.wrapper.Clear(ctx, s.spreadsheetID, []models.A1Range{s.scratchpadLocation})
	return err
}

// NewGoogleSheetKVStore creates an instance of the key-value store with the given configuration.
// It will also try to create the sheet, in case it does not exist yet.
func NewGoogleSheetKVStore(
	auth AuthClient,
	spreadsheetID string,
	sheetName string,
	config SheetKVStoreConfig,
) *SheetKVStore {
	wrapper, err := NewWrapper(auth)
	if err != nil {
		panic(fmt.Errorf("error creating sheets wrapper: %w", err))
	}

	scratchpadSheetName := sheetName + scratchpadSheetNameSuffix
	config = applyGoogleSheetKVStoreConfig(config)

	store := &SheetKVStore{
		wrapper:             wrapper,
		spreadsheetID:       spreadsheetID,
		sheetName:           sheetName,
		scratchpadSheetName: scratchpadSheetName,
		config:              config,
	}

	_ = ensureSheets(wrapper, spreadsheetID, sheetName)
	_ = ensureSheets(wrapper, spreadsheetID, scratchpadSheetName)

	scratchpadLocation, err := findScratchpadLocation(wrapper, spreadsheetID, scratchpadSheetName)
	if err != nil {
		panic(fmt.Errorf("error finding a scratchpad location in sheet %s: %w", scratchpadSheetName, err))
	}
	store.scratchpadLocation = scratchpadLocation

	return store
}

func applyGoogleSheetKVStoreConfig(config SheetKVStoreConfig) SheetKVStoreConfig {
	config.codec = codec.NewBasic()
	return config
}
