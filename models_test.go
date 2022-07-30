package freeleh

import (
	"context"
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
