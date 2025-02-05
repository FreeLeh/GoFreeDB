package google

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
)

// GoogleSheetRowStoreConfig defines a list of configurations that can be used to customise how the SheetRowStore works.
type GoogleSheetRowStoreConfig struct {
	// Columns defines the list of column names.
	// Note that the column ordering matters.
	// The column ordering will be used for arranging the real columns in Google Sheet.
	// Changing the column ordering in this config but not in Google Sheet will result in unexpected behaviour.
	Columns []string

	// ColumnsWithFormula defines the list of column names containing a Google Sheet formula.
	// Note that only string fields can have a formula.
	ColumnsWithFormula []string
}

func (c GoogleSheetRowStoreConfig) validate() error {
	if len(c.Columns) == 0 {
		return errors.New("columns must have at least one column")
	}
	if len(c.Columns) > maxColumn {
		return fmt.Errorf("you can only have up to %d columns", maxColumn)
	}
	return nil
}

// SheetRowStore encapsulates row store functionality on top of a Google Sheet.
type SheetRowStore struct {
	wrapper         sheetsWrapper
	spreadsheetID   string
	sheetName       string
	colsMapping     models.ColsMapping
	colsWithFormula *common.Set[string]
	config          GoogleSheetRowStoreConfig
}

// Select specifies which columns to return from the Google Sheet when querying and the output variable
// the data should be stored.
// You can think of this operation like a SQL SELECT statement (with limitations).
//
// If "columns" is an empty slice of string, then all columns will be returned.
// If a column is not found in the provided list of columns in `GoogleSheetRowStoreConfig.Columns`, that column will be ignored.
//
// "output" must be a pointer to a slice of a data type.
// The conversion from the Google Sheet data into the slice will be done using https://github.com/mitchellh/mapstructure.
//
// If you are providing a slice of structs into the "output" parameter, and you want to define the mapping between the
// column name with the field name, you should add a "db" struct tag.
//
//		// Without the `db` struct tag, the column name used will be "Name" and "Age".
//		type Person struct {
//	    	Name string `db:"name"`
//	    	Age int `db:"age"`
//		}
//
// Please note that calling Select() does not execute the query yet.
// Call SheetSelectStmt.Exec to actually execute the query.
func (s *SheetRowStore) Select(output interface{}, columns ...string) *SheetSelectStmt {
	return newSheetSelectStmt(s, output, columns)
}

// Insert specifies the rows to be inserted into the Google Sheet.
//
// The underlying data type of each row must be a struct or a pointer to a struct.
// Providing other data types will result in an error.
//
// By default, the column name will be following the struct field name (case-sensitive).
// If you want to map the struct field name into another name, you can add a "db" struct tag
// (see SheetRowStore.Select docs for more details).
//
// Please note that calling Insert() does not execute the insertion yet.
// Call SheetInsertStmt.Exec() to actually execute the insertion.
func (s *SheetRowStore) Insert(rows ...interface{}) *SheetInsertStmt {
	return newGoogleSheetInsertStmt(s, rows)
}

// Update specifies the new value for each of the targeted columns.
//
// The "colToValue" parameter specifies what value should be updated for which column.
// Each value in the map[string]interface{} is going to be JSON marshalled.
// If "colToValue" is empty, an error will be returned when SheetUpdateStmt.Exec() is called.
func (s *SheetRowStore) Update(colToValue map[string]interface{}) *SheetUpdateStmt {
	return newSheetUpdateStmt(s, colToValue)
}

// Delete prepares rows deletion operation.
//
// Please note that calling Delete() does not execute the deletion yet.
// Call SheetDeleteStmt.Exec() to actually execute the deletion.
func (s *SheetRowStore) Delete() *SheetDeleteStmt {
	return newSheetDeleteStmt(s)
}

// Count prepares rows counting operation.
//
// Please note that calling Count() does not execute the query yet.
// Call SheetCountStmt.Exec() to actually execute the query.
func (s *SheetRowStore) Count() *SheetCountStmt {
	return newSheetCountStmt(s)
}

// Close cleans up all held resources if any.
func (s *SheetRowStore) Close(_ context.Context) error {
	return nil
}

func (s *SheetRowStore) ensureHeaders() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	a1Range := models.NewA1Range(s.sheetName, defaultRowHeaderRange)
	if _, err := s.wrapper.Clear(
		ctx,
		s.spreadsheetID,
		[]models.A1Range{a1Range},
	); err != nil {
		return err
	}

	cols := make([]interface{}, len(s.config.Columns))
	for i := range s.config.Columns {
		cols[i] = s.config.Columns[i]
	}

	if _, err := s.wrapper.UpdateRows(
		ctx,
		s.spreadsheetID,
		models.NewA1Range(s.sheetName, defaultRowHeaderRange),
		[][]interface{}{cols},
	); err != nil {
		return err
	}
	return nil
}

// NewGoogleSheetRowStore creates an instance of the row based store with the given configuration.
// It will also try to create the sheet, in case it does not exist yet.
func NewGoogleSheetRowStore(
	auth AuthClient,
	spreadsheetID string,
	sheetName string,
	config GoogleSheetRowStoreConfig,
) *SheetRowStore {
	if err := config.validate(); err != nil {
		panic(err)
	}

	wrapper, err := NewWrapper(auth)
	if err != nil {
		panic(fmt.Errorf("error creating sheets wrapper: %w", err))
	}

	config = injectRIDCol(config)
	store := &SheetRowStore{
		wrapper:         wrapper,
		spreadsheetID:   spreadsheetID,
		sheetName:       sheetName,
		colsMapping:     models.GenerateColumnMapping(config.Columns),
		colsWithFormula: common.NewSet(config.ColumnsWithFormula),
		config:          config,
	}

	_ = ensureSheets(store.wrapper, store.spreadsheetID, store.sheetName)
	if err := store.ensureHeaders(); err != nil {
		panic(fmt.Errorf("error checking headers: %w", err))
	}
	return store
}

// The additional rowIdxCol column is needed to differentiate which row is truly empty and which one is not.
// Currently, we use this for detecting which rows are really empty for UPDATE without WHERE clause.
// Otherwise, it will always update all rows (instead of the non-empty rows only).
func injectRIDCol(config GoogleSheetRowStoreConfig) GoogleSheetRowStoreConfig {
	newCols := make([]string, 0, len(config.Columns)+1)
	newCols = append(newCols, rowIdxCol)
	newCols = append(newCols, config.Columns...)
	config.Columns = newCols

	return config
}
