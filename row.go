package freeleh

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FreeLeh/GoFreeLeh/internal/google/sheets"
)

type GoogleSheetRowStoreConfig struct {
	Columns []string
}

func (c GoogleSheetRowStoreConfig) validate() error {
	if len(c.Columns) == 0 {
		return errors.New("columns must have at least one column")
	}
	if len(c.Columns) > maxColumn {
		return errors.New("you can only have up to 1000 columns")
	}
	return nil
}

type GoogleSheetRowStore struct {
	wrapper             sheetsWrapper
	spreadsheetID       string
	sheetName           string
	scratchpadSheetName string
	scratchpadLocation  sheets.A1Range
	colsMapping         colsMapping
	config              GoogleSheetRowStoreConfig
}

func (s *GoogleSheetRowStore) Select(output interface{}, columns ...string) *googleSheetSelectStmt {
	return newGoogleSheetSelectStmt(s, output, columns)
}

// Insert will try to infer what is the type of each row and perform certain logic based on the type.
// For example, a struct will be converted into a map[string]interface{} and then into []interface{} (following the
// column mapping ordering).
//
// A few things to take note:
// - Only `struct` base type (including a pointer to a struct) is supported.
// - Each field name corresponds to the column name (case-sensitive).
// - The mapping between field name and column name can be changed by adding the struct field tag `db:"<col_name>"`.
func (s *GoogleSheetRowStore) Insert(rows ...interface{}) *googleSheetInsertStmt {
	return newGoogleSheetInsertStmt(s, rows)
}

// Update applies the given value for each column into the applicable rows.
// Note that in current version, we are not doing any data type validation for the given column-value pair.
// Each value in the `interface{}` is going to be JSON marshalled.
func (s *GoogleSheetRowStore) Update(colToValue map[string]interface{}) *googleSheetUpdateStmt {
	return newGoogleSheetUpdateStmt(s, colToValue)
}

func (s *GoogleSheetRowStore) Delete() *googleSheetDeleteStmt {
	return newGoogleSheetDeleteStmt(s)
}

func (s *GoogleSheetRowStore) Count() *googleSheetCountStmt {
	return newGoogleSheetCountStmt(s)
}

func (s *GoogleSheetRowStore) Close(ctx context.Context) error {
	_, err := s.wrapper.Clear(ctx, s.spreadsheetID, []string{s.scratchpadLocation.Original})
	return err
}

func (s *GoogleSheetRowStore) ensureHeaders() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if _, err := s.wrapper.Clear(
		ctx,
		s.spreadsheetID,
		[]string{getA1Range(s.sheetName, defaultRowHeaderRange)},
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
		getA1Range(s.sheetName, defaultRowHeaderRange),
		[][]interface{}{cols},
	); err != nil {
		return err
	}
	return nil
}

func NewGoogleSheetRowStore(
	auth sheets.AuthClient,
	spreadsheetID string,
	sheetName string,
	config GoogleSheetRowStoreConfig,
) *GoogleSheetRowStore {
	if err := config.validate(); err != nil {
		panic(err)
	}

	wrapper, err := sheets.NewWrapper(auth)
	if err != nil {
		panic(fmt.Errorf("error creating sheets wrapper: %w", err))
	}

	config = injectTimestampCol(config)
	scratchpadSheetName := sheetName + scratchpadSheetNameSuffix

	store := &GoogleSheetRowStore{
		wrapper:             wrapper,
		spreadsheetID:       spreadsheetID,
		sheetName:           sheetName,
		scratchpadSheetName: scratchpadSheetName,
		colsMapping:         generateColumnMapping(config.Columns),
		config:              config,
	}

	_ = ensureSheets(store.wrapper, store.spreadsheetID, store.sheetName)
	_ = ensureSheets(store.wrapper, store.spreadsheetID, store.scratchpadSheetName)

	if err := store.ensureHeaders(); err != nil {
		panic(fmt.Errorf("error checking headers: %w", err))
	}

	scratchpadLocation, err := findScratchpadLocation(store.wrapper, store.spreadsheetID, store.scratchpadSheetName)
	if err != nil {
		panic(fmt.Errorf("error finding a scratchpad location in sheet %s: %w", store.scratchpadSheetName, err))
	}
	store.scratchpadLocation = scratchpadLocation

	return store
}

// The additional _ts column is needed to differentiate which row is truly empty and which one is not.
// Currently, we use this for detecting which rows are really empty for UPDATE without WHERE clause.
// Otherwise, it will always update all rows (instead of the non-empty rows only).
func injectTimestampCol(config GoogleSheetRowStoreConfig) GoogleSheetRowStoreConfig {
	newCols := make([]string, 0, len(config.Columns)+1)
	newCols = append(newCols, rowIdxCol)
	newCols = append(newCols, config.Columns...)
	config.Columns = newCols

	return config
}
