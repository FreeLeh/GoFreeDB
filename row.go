package freedb

import (
	"github.com/FreeLeh/GoFreeDB/internal/google"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

type (
	GoogleSheetRowStore       = google.SheetRowStore
	GoogleSheetRowStoreConfig = google.GoogleSheetRowStoreConfig

	GoogleSheetSelectStmt = google.SheetSelectStmt
	GoogleSheetInsertStmt = google.SheetInsertStmt
	GoogleSheetUpdateStmt = google.SheetUpdateStmt
	GoogleSheetDeleteStmt = google.SheetDeleteStmt

	ColumnOrderBy = models.ColumnOrderBy
	OrderBy       = models.OrderBy
)

var (
	NewGoogleSheetRowStore = google.NewGoogleSheetRowStore

	OrderByAsc  = models.OrderByAsc
	OrderByDesc = models.OrderByDesc
)
