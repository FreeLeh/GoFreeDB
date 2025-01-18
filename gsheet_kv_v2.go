package freedb

import (
	"context"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
)

type googleSheetKVStoreV2Row struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

// GoogleSheetKVStoreV2Config defines a list of configurations that can be used to customise
// how the GoogleSheetKVStoreV2 works.
type GoogleSheetKVStoreV2Config struct {
	Mode  KVMode
	codec Codec
}

// GoogleSheetKVStoreV2 implements a key-value store using the row store abstraction.
type GoogleSheetKVStoreV2 struct {
	rowStore *GoogleSheetRowStore
	mode     KVMode
	codec    Codec
}

// Get retrieves the value associated with the given key.
func (s *GoogleSheetKVStoreV2) Get(ctx context.Context, key string) ([]byte, error) {
	var rows []googleSheetKVStoreV2Row
	var err error

	if s.mode == KVModeDefault {
		err = s.rowStore.Select(&rows, "value").
			Where("key = ?", key).
			Limit(1).
			Exec(ctx)
	} else {
		err = s.rowStore.Select(&rows, "value").
			Where("key = ?", key).
			OrderBy([]ColumnOrderBy{
				{Column: "_rid", OrderBy: OrderByDesc},
			}).
			Limit(1).
			Exec(ctx)
	}
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	value := rows[0].Value
	if value == "" {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	return s.codec.Decode(value)
}

// Set inserts or updates the key-value pair in the store.
func (s *GoogleSheetKVStoreV2) Set(ctx context.Context, key string, value []byte) error {
	encoded, err := s.codec.Encode(value)
	if err != nil {
		return err
	}

	row := googleSheetKVStoreV2Row{
		Key:   key,
		Value: encoded,
	}
	return s.rowStore.Insert(row).Exec(ctx)
}

// Delete removes the key from the store.
func (s *GoogleSheetKVStoreV2) Delete(ctx context.Context, key string) error {
	if s.mode == KVModeDefault {
		return s.rowStore.Delete().
			Where("key = ?", key).
			Exec(ctx)
	} else {
		return s.rowStore.Insert(googleSheetKVStoreV2Row{
			Key:   key,
			Value: "",
		}).Exec(ctx)
	}
}

// Close cleans up resources used by the store.
func (s *GoogleSheetKVStoreV2) Close(ctx context.Context) error {
	return s.rowStore.Close(ctx)
}

// NewGoogleSheetKVStoreV2 creates a new instance of the key-value store using row store.
// You cannot use this V2 store with the V1 store as the sheet format is different.
func NewGoogleSheetKVStoreV2(
	auth sheets.AuthClient,
	spreadsheetID string,
	sheetName string,
	config GoogleSheetKVStoreV2Config,
) *GoogleSheetKVStoreV2 {
	rowStore := NewGoogleSheetRowStore(
		auth,
		spreadsheetID,
		sheetName,
		GoogleSheetRowStoreConfig{
			Columns: []string{"key", "value"},
		},
	)

	config = applyGoogleSheetKVStoreV2Config(config)
	return &GoogleSheetKVStoreV2{
		rowStore: rowStore,
		mode:     config.Mode,
		codec:    config.codec,
	}
}

func applyGoogleSheetKVStoreV2Config(config GoogleSheetKVStoreV2Config) GoogleSheetKVStoreV2Config {
	config.codec = &basicCodec{}
	return config
}
