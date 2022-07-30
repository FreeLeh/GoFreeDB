package freeleh

import (
	"context"
	"fmt"
	"testing"

	"github.com/FreeLeh/GoFreeLeh/google/auth"
	"github.com/stretchr/testify/assert"
)

type testPerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	DOB  string `json:"dob"`
}

func TestGoogleSheetRowStore(t *testing.T) {
	spreadsheetID, authJSON, shouldRun := getIntegrationTestInfo()
	if !shouldRun {
		t.Skip("integration test should be run only in GitHub Actions")
	}
	sheetName := fmt.Sprintf("integration_row_%d", currentTimeMs())

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
		deleteSheet(t, db.wrapper, spreadsheetID, []string{db.sheetName, db.scratchpadSheetName})
		_ = db.Close(context.Background())
	}()

	var out []testPerson

	err = db.Select(&out, "name", "age").Exec(context.Background())
	assert.Nil(t, err)
	assert.Empty(t, out)

	err = db.RawInsert(
		[]interface{}{"name1", 10, "1-1-1999"},
		[]interface{}{"name2", 11, "1-1-2000"},
		[]interface{}{"name3", 12, "1-1-2001"},
	).Exec(context.Background())
	assert.Nil(t, err)

	err = db.Update(map[string]interface{}{"name": "name4"}).
		Where("age = ?", 10).
		Exec(context.Background())
	assert.Nil(t, err)

	expected := []testPerson{
		{"name2", 11, ""},
		{"name3", 12, ""},
	}
	err = db.Select(&out, "name", "age").
		Where("name = ? OR name = ?", "name2", "name3").
		OrderBy(map[string]OrderBy{"name": OrderByAsc}).
		Limit(2).
		Exec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expected, out)

	err = db.Delete().Where("name = ?", "name4").Exec(context.Background())
	assert.Nil(t, err)
}
