package store

import (
	"context"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"strconv"
	"testing"

	"github.com/FreeLeh/GoFreeDB/google/auth"
	"github.com/stretchr/testify/assert"
)

type testPerson struct {
	Name string `json:"name" db:"name"`
	Age  int64  `json:"age" db:"age"`
	DOB  string `json:"dob" db:"dob"`
}

func TestGoogleSheetRowStore_Integration(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_row_%d", common.CurrentTimeMs())

	googleAuth, err := auth.NewServiceFromJSON([]byte(authJSON), auth.GoogleSheetsReadWrite, auth.ServiceConfig{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	db := NewGoogleSheetRowStore(
		googleAuth,
		spreadsheetID,
		sheetName,
		GoogleSheetRowStoreConfig{Columns: []string{"name", "age", "dob"}},
	)
	defer func() {
		deleteSheet(t, db.wrapper, spreadsheetID, []string{db.sheetName})
		_ = db.Close(context.Background())
	}()

	var out []testPerson

	err = db.Select(&out, "name", "age").Offset(10).Limit(10).Exec(context.Background())
	assert.Nil(t, err)
	assert.Empty(t, out)

	err = db.Insert(
		testPerson{"name1", 10, "1999-01-01"},
		testPerson{"name2", 11, "2000-01-01"},
	).Exec(context.Background())
	assert.Nil(t, err)

	// Nil type
	err = db.Insert(nil).Exec(context.Background())
	assert.NotNil(t, err)

	err = db.Insert(testPerson{
		Name: "name3",
		Age:  9007199254740992,
		DOB:  "2001-01-01",
	}).Exec(context.Background())
	assert.Nil(t, err)

	err = db.Update(map[string]interface{}{"name": "name4"}).
		Where("age = ?", 10).
		Exec(context.Background())
	assert.Nil(t, err)

	expected := []testPerson{
		{"name2", 11, "2000-01-01"},
		{"name3", 9007199254740992, "2001-01-01"},
	}
	err = db.Select(&out, "name", "age", "dob").
		Where("name = ? OR name = ?", "name2", "name3").
		OrderBy([]models.ColumnOrderBy{{"name", models.OrderByAsc}}).
		Limit(2).
		Exec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, out)

	count, err := db.Count().
		Where("name = ? OR name = ?", "name2", "name3").
		Exec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), count)

	err = db.Delete().Where("name = ?", "name4").Exec(context.Background())
	assert.Nil(t, err)
}

func TestGoogleSheetRowStore_Integration_EdgeCases(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_row_%d", common.CurrentTimeMs())

	googleAuth, err := auth.NewServiceFromJSON([]byte(authJSON), auth.GoogleSheetsReadWrite, auth.ServiceConfig{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	db := NewGoogleSheetRowStore(
		googleAuth,
		spreadsheetID,
		sheetName,
		GoogleSheetRowStoreConfig{Columns: []string{"name", "age", "dob"}},
	)
	defer func() {
		deleteSheet(t, db.wrapper, spreadsheetID, []string{db.sheetName})
		_ = db.Close(context.Background())
	}()

	// Non-struct types
	err = db.Insert([]interface{}{"name3", 12, "2001-01-01"}).Exec(context.Background())
	assert.NotNil(t, err)

	// IEEE 754 unsafe integer
	err = db.Insert([]interface{}{"name3", 9007199254740993, "2001-01-01"}).Exec(context.Background())
	assert.NotNil(t, err)

	// IEEE 754 unsafe integer
	err = db.Insert(
		testPerson{"name1", 10, "1999-01-01"},
		testPerson{"name2", 11, "2000-01-01"},
	).Exec(context.Background())
	assert.Nil(t, err)

	err = db.Update(map[string]interface{}{"name": "name4", "age": int64(9007199254740993)}).
		Exec(context.Background())
	assert.NotNil(t, err)
}

func TestInjectTimestampCol(t *testing.T) {
	result := injectTimestampCol(GoogleSheetRowStoreConfig{Columns: []string{"col1", "col2"}})
	assert.Equal(t, GoogleSheetRowStoreConfig{Columns: []string{rowIdxCol, "col1", "col2"}}, result)
}

func TestGoogleSheetRowStoreConfig_validate(t *testing.T) {
	t.Run("empty_columns", func(t *testing.T) {
		conf := GoogleSheetRowStoreConfig{Columns: []string{}}
		assert.NotNil(t, conf.validate())
	})

	t.Run("too_many_columns", func(t *testing.T) {
		columns := make([]string, 0)
		for i := 0; i < 27; i++ {
			columns = append(columns, strconv.FormatInt(int64(i), 10))
		}

		conf := GoogleSheetRowStoreConfig{Columns: columns}
		assert.NotNil(t, conf.validate())
	})

	t.Run("no_error", func(t *testing.T) {
		columns := make([]string, 0)
		for i := 0; i < 10; i++ {
			columns = append(columns, strconv.FormatInt(int64(i), 10))
		}

		conf := GoogleSheetRowStoreConfig{Columns: columns}
		assert.Nil(t, conf.validate())
	})
}
