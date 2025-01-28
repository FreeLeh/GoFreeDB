package store

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FreeLeh/GoFreeDB/internal/common"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"github.com/FreeLeh/GoFreeDB/lark/auth"
)

type testPerson struct {
	Name string `json:"name" db:"name"`
	Age  int64  `json:"age" db:"age"`
	DOB  string `json:"dob" db:"dob"`
}

func TestGoogleSheetRowStore_Integration(t *testing.T) {
	spreadsheetID, authCfg, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_row_%d", common.CurrentTimeMs())

	larkAuth, err := auth.NewTenantService(authCfg.AppID, authCfg.AppSecret, auth.Config{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	db := NewLarkSheetRowStore(
		larkAuth,
		spreadsheetID,
		sheetName,
		LarkSheetRowStoreConfig{Columns: []string{"name", "age", "dob"}},
	)
	defer func() {
		time.Sleep(time.Second)
		deleteSheet(t, db.wrapper, spreadsheetID, []string{db.sheetName})
		_ = db.Close(context.Background())
	}()

	var out []testPerson

	time.Sleep(time.Second)
	err = db.Select(&out, "name", "age").Offset(10).Limit(10).Exec(context.Background())
	assert.Nil(t, err)
	assert.Empty(t, out)

	time.Sleep(time.Second)
	err = db.Insert(
		testPerson{"name1", 10, "1999-01-01"},
		testPerson{"name2", 11, "2000-01-01"},
	).Exec(context.Background())
	assert.Nil(t, err)

	// Nil type
	time.Sleep(time.Second)
	err = db.Insert(nil).Exec(context.Background())
	assert.NotNil(t, err)

	time.Sleep(time.Second)
	err = db.Insert(testPerson{
		Name: "name3",
		Age:  9007199254740992,
		DOB:  "2001-01-01",
	}).Exec(context.Background())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = db.Update(map[string]interface{}{"name": "name4"}).
		Where("age = ?", 10).
		Exec(context.Background())
	assert.Nil(t, err)

	expected := []testPerson{
		{"name2", 11, "2000-01-01"},
		{"name3", 9007199254740992, "2001-01-01"},
	}

	time.Sleep(time.Second)
	err = db.Select(&out, "name", "age", "dob").
		Where("name = ? OR name = ?", "name2", "name3").
		OrderBy([]models.ColumnOrderBy{{"name", models.OrderByAsc}}).
		Limit(2).
		Exec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, out)

	time.Sleep(time.Second)
	count, err := db.Count().
		Where("name = ? OR name = ?", "name2", "name3").
		Exec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), count)

	time.Sleep(time.Second)
	err = db.Delete().Where("name = ?", "name4").Exec(context.Background())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = db.Insert(
		testPerson{"name100", 10, "1999-01-01"},
		testPerson{"name200", 11, "2000-01-01"},
		testPerson{"name300", 12, "2001-01-01"},
	).Exec(context.Background())
	assert.Nil(t, err)

	var out2 []testPerson
	expected = []testPerson{
		{"name100", 10, "1999-01-01"},
		{"name2", 11, "2000-01-01"},
		{"name200", 11, "2000-01-01"},
		{"name3", 9007199254740992, "2001-01-01"},
		{"name300", 12, "2001-01-01"},
	}

	time.Sleep(time.Second)
	err = db.Select(&out2, "name", "age", "dob").
		OrderBy([]models.ColumnOrderBy{{"name", models.OrderByAsc}}).
		Exec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, out2)
}

func TestGoogleSheetRowStore_Integration_EdgeCases(t *testing.T) {
	spreadsheetID, authCfg, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_row_%d", common.CurrentTimeMs())

	larkAuth, err := auth.NewTenantService(authCfg.AppID, authCfg.AppSecret, auth.Config{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	db := NewLarkSheetRowStore(
		larkAuth,
		spreadsheetID,
		sheetName,
		LarkSheetRowStoreConfig{Columns: []string{"name", "age", "dob"}},
	)
	defer func() {
		time.Sleep(time.Second)
		deleteSheet(t, db.wrapper, spreadsheetID, []string{db.sheetName})
		_ = db.Close(context.Background())
	}()

	// Non-struct types
	time.Sleep(time.Second)
	err = db.Insert([]interface{}{"name3", 12, "2001-01-01"}).Exec(context.Background())
	assert.NotNil(t, err)

	// IEEE 754 unsafe integer
	time.Sleep(time.Second)
	err = db.Insert([]interface{}{"name3", 9007199254740993, "2001-01-01"}).Exec(context.Background())
	assert.NotNil(t, err)

	// IEEE 754 unsafe integer
	time.Sleep(time.Second)
	err = db.Insert(
		testPerson{"name1", 10, "1999-01-01"},
		testPerson{"name2", 11, "2000-01-01"},
	).Exec(context.Background())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = db.Update(map[string]interface{}{"name": "name4", "age": int64(9007199254740993)}).
		Exec(context.Background())
	assert.NotNil(t, err)
}

type formulaWriteModel struct {
	Value string `json:"value" db:"value"`
}

type formulaReadModel struct {
	Value int `json:"value" db:"value"`
}

func TestGoogleSheetRowStore_Formula(t *testing.T) {
	spreadsheetID, authCfg, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_row_%d", common.CurrentTimeMs())

	larkAuth, err := auth.NewTenantService(authCfg.AppID, authCfg.AppSecret, auth.Config{})
	if err != nil {
		t.Fatalf("error when instantiating google auth: %s", err)
	}

	db := NewLarkSheetRowStore(
		larkAuth,
		spreadsheetID,
		sheetName,
		LarkSheetRowStoreConfig{
			Columns:            []string{"value"},
			ColumnsWithFormula: []string{"value"},
		},
	)
	defer func() {
		time.Sleep(time.Second)
		deleteSheet(t, db.wrapper, spreadsheetID, []string{db.sheetName})
		_ = db.Close(context.Background())
	}()

	var out []formulaReadModel

	time.Sleep(time.Second)
	err = db.Insert(formulaWriteModel{Value: "=ROW()-1"}).Exec(context.Background())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = db.Select(&out).Exec(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []formulaReadModel{{Value: 1}}, out)

	time.Sleep(time.Second)
	err = db.Update(map[string]interface{}{"value": "=ROW()"}).Exec(context.Background())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = db.Select(&out).Exec(context.Background())
	assert.Nil(t, err)
	assert.ElementsMatch(t, []formulaReadModel{{Value: 2}}, out)
}

func TestInjectTimestampCol(t *testing.T) {
	result := injectRIDCol(LarkSheetRowStoreConfig{Columns: []string{"col1", "col2"}})
	assert.Equal(t, LarkSheetRowStoreConfig{Columns: []string{rowIdxCol, "col1", "col2"}}, result)
}

func TestLarkSheetRowStoreConfig_validate(t *testing.T) {
	t.Run("empty_columns", func(t *testing.T) {
		conf := LarkSheetRowStoreConfig{Columns: []string{}}
		assert.NotNil(t, conf.validate())
	})

	t.Run("too_many_columns", func(t *testing.T) {
		columns := make([]string, 0)
		for i := 0; i < 27; i++ {
			columns = append(columns, strconv.FormatInt(int64(i), 10))
		}

		conf := LarkSheetRowStoreConfig{Columns: columns}
		assert.NotNil(t, conf.validate())
	})

	t.Run("no_error", func(t *testing.T) {
		columns := make([]string, 0)
		for i := 0; i < 10; i++ {
			columns = append(columns, strconv.FormatInt(int64(i), 10))
		}

		conf := LarkSheetRowStoreConfig{Columns: columns}
		assert.Nil(t, conf.validate())
	})
}
