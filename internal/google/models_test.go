package google

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getIntegrationTestInfo() (string, string, bool) {
	spreadsheetID := os.Getenv("INTEGRATION_TEST_SPREADSHEET_ID")
	authJSON := os.Getenv("INTEGRATION_TEST_AUTH_JSON")
	_, isGithubActions := os.LookupEnv("GITHUB_ACTIONS")
	return spreadsheetID, authJSON, isGithubActions && spreadsheetID != "" && authJSON != ""
}

func deleteSheet(t *testing.T, wrapper sheetsWrapper, spreadsheetID string, sheetNames []string) {
	sheetNameToID, err := wrapper.GetSheetNameToID(context.Background(), spreadsheetID)
	if err != nil {
		t.Fatalf("failed getting mapping of sheet names to IDs: %s", err)
	}

	sheetIDs := make([]int64, 0, len(sheetNames))
	for _, sheetName := range sheetNames {
		id, ok := sheetNameToID[sheetName]
		if !ok {
			t.Fatalf("sheet ID for the given name is not found")
		}
		sheetIDs = append(sheetIDs, id)
	}

	if err := wrapper.DeleteSheets(context.Background(), spreadsheetID, sheetIDs); err != nil {
		t.Logf("failed deleting sheets: %s", err)
	}
}

func TestRawQueryRowsResult_toQueryRowsResult(t *testing.T) {
	t.Run("empty_rows", func(t *testing.T) {
		r := rawQueryRowsResult{
			Table: rawQueryRowsResultTable{
				Cols: []rawQueryRowsResultColumn{
					{ID: "A", Type: "number"},
					{ID: "B", Type: "string"},
				},
				Rows: []rawQueryRowsResultRow{},
			},
		}

		expected := QueryRowsResult{Rows: make([][]interface{}, 0)}

		result, err := r.toQueryRowsResult()
		assert.Nil(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("few_rows", func(t *testing.T) {
		r := rawQueryRowsResult{
			Table: rawQueryRowsResultTable{
				Cols: []rawQueryRowsResultColumn{
					{ID: "A", Type: "number"},
					{ID: "B", Type: "string"},
					{ID: "C", Type: "boolean"},
				},
				Rows: []rawQueryRowsResultRow{
					{
						[]rawQueryRowsResultCell{
							{Value: 123.0, Raw: "123"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "true"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 456.0, Raw: "456"},
							{Value: "blah2", Raw: "blah2"},
							{Value: false, Raw: "FALSE"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 123.1, Raw: "123.1"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "TRUE"},
						},
					},
				},
			},
		}

		expected := QueryRowsResult{
			Rows: [][]interface{}{
				{float64(123), "blah", true},
				{float64(456), "blah2", false},
				{123.1, "blah", true},
			},
		}

		result, err := r.toQueryRowsResult()
		assert.Nil(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("unexpected_type", func(t *testing.T) {
		r := rawQueryRowsResult{
			Table: rawQueryRowsResultTable{
				Cols: []rawQueryRowsResultColumn{
					{ID: "A", Type: "number"},
					{ID: "B", Type: "string"},
					{ID: "C", Type: "something"},
				},
				Rows: []rawQueryRowsResultRow{
					{
						[]rawQueryRowsResultCell{
							{Value: 123.0, Raw: "123"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "true"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 456.0, Raw: "456"},
							{Value: "blah2", Raw: "blah2"},
							{Value: false, Raw: "FALSE"},
						},
					},
					{
						[]rawQueryRowsResultCell{
							{Value: 123.1, Raw: "123.1"},
							{Value: "blah", Raw: "blah"},
							{Value: true, Raw: "TRUE"},
						},
					},
				},
			},
		}

		result, err := r.toQueryRowsResult()
		assert.Equal(t, QueryRowsResult{}, result)
		assert.NotNil(t, err)
	})
}
