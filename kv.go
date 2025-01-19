package freedb

import (
	"github.com/FreeLeh/GoFreeDB/internal/google/store"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type (
	GoogleSheetKVStore       = store.GoogleSheetKVStore
	GoogleSheetKVStoreConfig = store.GoogleSheetKVStoreConfig
	KVMode                   = models.KVMode

	GoogleSheetKVStoreV2       = store.GoogleSheetKVStoreV2
	GoogleSheetKVStoreV2Config = store.GoogleSheetKVStoreV2Config
)

var (
	NewGoogleSheetKVStore   = store.NewGoogleSheetKVStore
	NewGoogleSheetKVStoreV2 = store.NewGoogleSheetKVStoreV2

	KVModeDefault    = models.KVModeDefault
	KVModeAppendOnly = models.KVModeAppendOnly
)
