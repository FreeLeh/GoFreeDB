package store

import (
	"context"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"testing"

	"github.com/FreeLeh/GoFreeDB/google/auth"
	"github.com/stretchr/testify/assert"
)

func TestGoogleSheetKVStoreV2_AppendOnly_Integration(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_kv_v2_append_only_%d", common.CurrentTimeMs())

	googleAuth, err := auth.NewServiceFromJSON([]byte(authJSON), auth.GoogleSheetsReadWrite, auth.ServiceConfig{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	kv := NewGoogleSheetKVStoreV2(
		googleAuth,
		spreadsheetID,
		sheetName,
		GoogleSheetKVStoreV2Config{Mode: models.KVModeAppendOnly},
	)
	defer func() {
		deleteSheet(t, kv.rowStore.wrapper, spreadsheetID, []string{kv.rowStore.sheetName})
		_ = kv.Close(context.Background())
	}()

	value, err := kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, models.ErrKeyNotFound)

	err = kv.Set(context.Background(), "k1", []byte("test"))
	assert.Nil(t, err)

	value, err = kv.Get(context.Background(), "k1")
	assert.Equal(t, []byte("test"), value)
	assert.Nil(t, err)

	err = kv.Delete(context.Background(), "k1")
	assert.Nil(t, err)

	value, err = kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, models.ErrKeyNotFound)
}

func TestNewGoogleSheetKVStoreV2_Default_Integration(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_kv_v2_default_%d", common.CurrentTimeMs())

	googleAuth, err := auth.NewServiceFromJSON([]byte(authJSON), auth.GoogleSheetsReadWrite, auth.ServiceConfig{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	kv := NewGoogleSheetKVStoreV2(
		googleAuth,
		spreadsheetID,
		sheetName,
		GoogleSheetKVStoreV2Config{Mode: models.KVModeDefault},
	)
	defer func() {
		deleteSheet(t, kv.rowStore.wrapper, spreadsheetID, []string{kv.rowStore.sheetName})
		_ = kv.Close(context.Background())
	}()

	value, err := kv.Get(context.Background(), "k1")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, models.ErrKeyNotFound)

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
	assert.ErrorIs(t, err, models.ErrKeyNotFound)
}
