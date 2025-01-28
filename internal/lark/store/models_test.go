package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

type larkAuthConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_token"`
}

func getIntegrationTestInfo() (string, larkAuthConfig, bool) {
	spreadsheetID := os.Getenv("INTEGRATION_TEST_LARK_SPREADSHEET_TOKEN")
	authJSON := os.Getenv("INTEGRATION_TEST_LARK_AUTH_JSON")
	_, isGithubActions := os.LookupEnv("GITHUB_ACTIONS")

	var larkAuthConfig larkAuthConfig
	if err := json.Unmarshal([]byte(authJSON), &larkAuthConfig); err != nil {
		panic(err)
	}

	return spreadsheetID, larkAuthConfig, isGithubActions && spreadsheetID != "" && authJSON != ""
}

func deleteSheet(t *testing.T, wrapper sheetsWrapper, spreadsheetToken string, sheetNames []string) {
	mapping, err := getSheetIDs(wrapper, spreadsheetToken)
	if err != nil {
		t.Fatalf("failed getting mapping of sheet names to IDs: %s", err)
	}

	sheetIDs := make([]string, 0, len(sheetNames))
	for _, sheetName := range sheetNames {
		id, ok := mapping[sheetName]
		if !ok {
			t.Fatalf("sheet ID for the given name is not found")
		}
		sheetIDs = append(sheetIDs, id)
	}

	if err := wrapper.DeleteSheets(context.Background(), spreadsheetToken, sheetIDs); err != nil {
		t.Logf("failed deleting sheets: %s", err)
	}
}
