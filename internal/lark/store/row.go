package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/lark/sheets"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"time"
)

// LarkSheetRowStoreConfig defines a list of configurations that can be used to customise how the LarkSheetRowStore works.
type LarkSheetRowStoreConfig struct {
	// Columns defines the list of column names.
	// Note that the column ordering matters.
	// The column ordering will be used for arranging the real columns in Lark Sheets.
	// Changing the column ordering in this config but not in Lark Sheets will result in unexpected behaviour.
	Columns []string

	// ColumnsWithFormula defines the list of column names containing a Lark Sheet formula.
	// Note that only string fields can have a formula.
	ColumnsWithFormula []string
}

func (c LarkSheetRowStoreConfig) validate() error {
	if len(c.Columns) == 0 {
		return errors.New("columns must have at least one column")
	}
	if len(c.Columns) > maxColumn {
		return fmt.Errorf("you can only have up to %d columns", maxColumn)
	}
	return nil
}

type LarkSheetRowStore struct {
	wrapper             sheetsWrapper
	spreadsheetToken    string
	sheetName           string
	sheetID             string
	scratchpadSheetName string
	scratchpadSheetID   string
	scratchpadLocation  models.A1Range
	colsMapping         models.ColsMapping
	colsWithFormula     *common.Set[string]
	config              LarkSheetRowStoreConfig
}

// Select specifies which columns to return from the Lark Sheets when querying and the output variable
// the data should be stored.
// You can think of this operation like a SQL SELECT statement (with limitations).
//
// If "columns" is an empty slice of string, then all columns will be returned.
// If a column is not found in the provided list of columns in `LarkSheetRowStoreConfig.Columns`, that column will be ignored.
//
// "output" must be a pointer to a slice of a data type.
// The conversion from the Lark Sheets data into the slice will be done using https://github.com/mitchellh/mapstructure.
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
// Call LarkSheetSelectStmt.Exec to actually execute the query.
func (s *LarkSheetRowStore) Select(output interface{}, columns ...string) *LarkSheetSelectStmt {
	return newLarkSheetSelectStmt(s, output, columns)
}

// Insert specifies the rows to be inserted into the Lark Sheets.
//
// The underlying data type of each row must be a struct or a pointer to a struct.
// Providing other data types will result in an error.
//
// By default, the column name will be following the struct field name (case-sensitive).
// If you want to map the struct field name into another name, you can add a "db" struct tag
// (see LarkSheetRowStore.Select docs for more details).
//
// Please note that calling Insert() does not execute the insertion yet.
// Call LarkSheetInsertStmt.Exec() to actually execute the insertion.
func (s *LarkSheetRowStore) Insert(rows ...interface{}) *LarkSheetInsertStmt {
	return newLarkSheetInsertStmt(s, rows)
}

// Update specifies the new value for each of the targeted columns.
//
// The "colToValue" parameter specifies what value should be updated for which column.
// Each value in the map[string]interface{} is going to be JSON marshalled.
// If "colToValue" is empty, an error will be returned when LarkSheetUpdateStmt.Exec() is called.
func (s *LarkSheetRowStore) Update(colToValue map[string]interface{}) *LarkSheetUpdateStmt {
	return newLarkSheetUpdateStmt(s, colToValue)
}

// Delete prepares rows deletion operation.
//
// Please note that calling Delete() does not execute the deletion yet.
// Call LarkSheetDeleteStmt.Exec() to actually execute the deletion.
func (s *LarkSheetRowStore) Delete() *LarkSheetDeleteStmt {
	return newLarkSheetDeleteStmt(s)
}

// Count prepares rows counting operation.
//
// Please note that calling Count() does not execute the query yet.
// Call GoogleSheetCountStmt.Exec() to actually execute the query.
func (s *LarkSheetRowStore) Count() *GoogleSheetCountStmt {
	return newGoogleSheetCountStmt(s)
}

// Close cleans up all held resources if any.
func (s *LarkSheetRowStore) Close(_ context.Context) error {
	return nil
}

func (s *LarkSheetRowStore) ensureHeaders() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	a1Range := models.NewA1Range(s.sheetName, defaultRowHeaderRange)
	if err := s.wrapper.Clear(
		ctx,
		s.spreadsheetToken,
		[]models.A1Range{a1Range},
	); err != nil {
		return err
	}

	cols := make([]interface{}, len(s.config.Columns))
	for i := range s.config.Columns {
		cols[i] = s.config.Columns[i]
	}

	if _, err := s.wrapper.BatchUpdateRows(
		ctx,
		s.spreadsheetToken,
		[]sheets.BatchUpdateRowsRequest{
			{
				A1Range: models.NewA1Range(s.sheetName, defaultRowHeaderRange),
				Values:  [][]interface{}{cols},
			},
		},
	); err != nil {
		return err
	}
	return nil
}

func (s *LarkSheetRowStore) validateAndFillSheetIDs() error {
	mapping, err := getSheetIDs(s.wrapper, s.spreadsheetToken)
	if err != nil {
		return err
	}

	mainSheetID, ok := mapping[s.sheetName]
	if !ok {
		return fmt.Errorf("sheet %s not found", s.sheetName)
	}

	scratchpadSheetID, ok := mapping[s.scratchpadSheetName]
	if !ok {
		return fmt.Errorf("scratchpad sheet %s not found", s.scratchpadSheetName)
	}

	s.sheetID = mainSheetID
	s.scratchpadSheetID = scratchpadSheetID

	return nil
}

// NewLarkSheetRowStore creates an instance of the row based store with the given configuration.
// It will also try to create the sheet, in case it does not exist yet.
func NewLarkSheetRowStore(
	accessTokenGetter sheets.AccessTokenGetter,
	spreadsheetToken string,
	sheetName string,
	config LarkSheetRowStoreConfig,
) *LarkSheetRowStore {
	if err := config.validate(); err != nil {
		panic(err)
	}

	scratchpadSheetName := sheetName + scratchpadSheetNameSuffix
	wrapper := sheets.NewWrapper(accessTokenGetter)
	config = injectRIDCol(config)

	store := &LarkSheetRowStore{
		wrapper:             wrapper,
		spreadsheetToken:    spreadsheetToken,
		sheetName:           sheetName,
		scratchpadSheetName: scratchpadSheetName,
		colsMapping:         models.GenerateColumnMapping(config.Columns),
		colsWithFormula:     common.NewSet(config.ColumnsWithFormula),
		config:              config,
	}

	_ = ensureSheets(store.wrapper, store.spreadsheetToken, store.sheetName)
	_ = ensureSheets(store.wrapper, store.spreadsheetToken, store.scratchpadSheetName)

	if err := store.validateAndFillSheetIDs(); err != nil {
		panic(fmt.Errorf("error validate and fill sheet IDs: %w", err))
	}
	if err := store.ensureHeaders(); err != nil {
		panic(fmt.Errorf("error checking headers: %w", err))
	}

	scratchpadLocation, err := findScratchpadLocation(store.wrapper, store.spreadsheetToken, store.scratchpadSheetName)
	if err != nil {
		panic(fmt.Errorf("error finding a scratchpad location in sheet %s: %w", store.scratchpadSheetName, err))
	}
	store.scratchpadLocation = scratchpadLocation

	return store
}

// The additional rowIdxCol column is needed to differentiate which row is truly empty and which one is not.
// Currently, we use this for detecting which rows are really empty for UPDATE without WHERE clause.
// Otherwise, it will always update all rows (instead of the non-empty rows only).
func injectRIDCol(config LarkSheetRowStoreConfig) LarkSheetRowStoreConfig {
	newCols := make([]string, 0, len(config.Columns)+1)
	newCols = append(newCols, rowIdxCol)
	newCols = append(newCols, config.Columns...)
	config.Columns = newCols

	return config
}
