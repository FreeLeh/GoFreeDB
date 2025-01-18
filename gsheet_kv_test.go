package freedb

import (
	"context"
	"fmt"
	"testing"

	"github.com/FreeLeh/GoFreeDB/google/auth"
	"github.com/stretchr/testify/assert"
)

func TestGoogleSheetKVStore_AppendOnly_Integration(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_kv_append_only_%d", currentTimeMs())

	googleAuth, err := auth.NewServiceFromJSON([]byte(authJSON), auth.GoogleSheetsReadWrite, auth.ServiceConfig{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	kv := NewGoogleSheetKVStore(
		googleAuth,
		spreadsheetID,
		sheetName,
		GoogleSheetKVStoreConfig{Mode: KVModeAppendOnly},
	)
	defer func() {
		deleteSheet(t, kv.wrapper, spreadsheetID, []string{kv.sheetName, kv.scratchpadSheetName})
		_ = kv.Close(context.Background())
	}()

	value, err := kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, ErrKeyNotFound)

	err = kv.Set(context.Background(), "k1", []byte("test"))
	assert.Nil(t, err)

	value, err = kv.Get(context.Background(), "k1")
	assert.Equal(t, []byte("test"), value)
	assert.Nil(t, err)

	err = kv.Delete(context.Background(), "k1")
	assert.Nil(t, err)

	value, err = kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestNewGoogleSheetKVStore_Default_Integration(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_kv_default_%d", currentTimeMs())

	googleAuth, err := auth.NewServiceFromJSON([]byte(authJSON), auth.GoogleSheetsReadWrite, auth.ServiceConfig{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	kv := NewGoogleSheetKVStore(
		googleAuth,
		spreadsheetID,
		sheetName,
		GoogleSheetKVStoreConfig{Mode: KVModeDefault},
	)
	defer func() {
		deleteSheet(t, kv.wrapper, spreadsheetID, []string{kv.sheetName, kv.scratchpadSheetName})
		_ = kv.Close(context.Background())
	}()

	value, err := kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, ErrKeyNotFound)

	err = kv.Set(context.Background(), "k1", []byte("test"))
	assert.Nil(t, err)

	value, err = kv.Get(context.Background(), "k1")
	assert.Equal(t, []byte("test"), value)
	assert.Nil(t, err)

	err = kv.Set(context.Background(), "k1", []byte("test2"))
	assert.Nil(t, err)

	err = kv.Delete(context.Background(), "k1")
	assert.Nil(t, err)

	value, err = kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}
