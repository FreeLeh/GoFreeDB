package freedb

import (
	"github.com/FreeLeh/GoFreeDB/internal/google/store"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type (
	GoogleSheetRowStore       = store.GoogleSheetRowStore
	GoogleSheetRowStoreConfig = store.GoogleSheetRowStoreConfig

	GoogleSheetSelectStmt = store.GoogleSheetSelectStmt
	GoogleSheetInsertStmt = store.GoogleSheetInsertStmt
	GoogleSheetUpdateStmt = store.GoogleSheetUpdateStmt
	GoogleSheetDeleteStmt = store.GoogleSheetDeleteStmt

	ColumnOrderBy = models.ColumnOrderBy
	OrderBy       = models.OrderBy
)

var (
	NewGoogleSheetRowStore = store.NewGoogleSheetRowStore

	OrderByAsc  = models.OrderByAsc
	OrderByDesc = models.OrderByDesc
)
