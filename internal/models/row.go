package models

// OrderBy defines the type of column ordering used for GoogleSheetRowStore.Select().
type OrderBy string

const (
	OrderByAsc  OrderBy = "ASC"
	OrderByDesc OrderBy = "DESC"
)

// ColumnOrderBy defines what ordering is required for a particular column.
// This is used for GoogleSheetRowStore.Select().
type ColumnOrderBy struct {
	Column  string
	OrderBy OrderBy
}
