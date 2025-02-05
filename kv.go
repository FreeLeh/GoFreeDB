package freedb

import (
	"github.com/FreeLeh/GoFreeDB/internal/google"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type (
	GoogleSheetKVStore       = google.SheetKVStore
	GoogleSheetKVStoreConfig = google.SheetKVStoreConfig
	KVMode                   = models.KVMode

	GoogleSheetKVStoreV2       = google.SheetKVStoreV2
	GoogleSheetKVStoreV2Config = google.SheetKVStoreV2Config
)

var (
	NewGoogleSheetKVStore   = google.NewGoogleSheetKVStore
	NewGoogleSheetKVStoreV2 = google.NewGoogleSheetKVStoreV2

	KVModeDefault    = models.KVModeDefault
	KVModeAppendOnly = models.KVModeAppendOnly
)
