package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	appID         = "cli_a7bf762734f89009"
	appSecret     = "XRwcNQT0sfUoBjl1g2R1qhlyjbKTmJAU"
	tokenURL      = "https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal"
	appendDataURL = "https://open.larksuite.com/open-apis/sheets/v2/spreadsheets/%s/values_append?insertDataOption=OVERWRITE"
)

func getAccessToken() (string, error) {
	// Request body for authentication
	body := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	bodyJSON, _ := json.Marshal(body)

	// Make POST request to get access token
	resp, err := http.Post(tokenURL, "application/json", bytes.NewBuffer(bodyJSON))
	if err != nil {
		return "", fmt.Errorf("failed to request access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %v", resp.Status)
	}

	// Parse response
	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse access token response: %v", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("error fetching token: %v", result.Msg)
	}

	return result.TenantAccessToken, nil
}

func appendData(accessToken, spreadsheetToken string, rangeName string, values [][]string) error {
	// Prepare API URL with spreadsheet token
	url := fmt.Sprintf(appendDataURL, spreadsheetToken)

	// Request body for appending data
	body := map[string]interface{}{
		"valueRange": map[string]interface{}{
			"range":  rangeName,
			"values": values,
		},
	}
	bodyJSON, _ := json.Marshal(body)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send append data request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("append data failed: %s", string(bodyBytes))
	}

	return nil
}

func main() {
	// Step 1: Get access token
	accessToken, err := getAccessToken()
	if err != nil {
		fmt.Printf("Error getting access token: %v\n", err)
		return
	}

	// Step 2: Append data
	spreadsheetToken := "RQYusDj9BhtUQUtXyTAuTHristf"
	rangeName := "070fc5" // Adjust the range name as needed
	values := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "25", "New York"},
		{"Bob", "30", "San Francisco"},
	}

	err = appendData(accessToken, spreadsheetToken, rangeName, values)
	if err != nil {
		fmt.Printf("Error appending data: %v\n", err)
		return
	}

	fmt.Println("Data appended successfully!")
}
