package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const accessTokenURL = "https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal"

type accessTokenResponse struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
}

type api struct {
	client *http.Client

	accessTokenURL string
}

func (api *api) TenantAccessToken(
	ctx context.Context,
	appID string,
	appSecret string,
) (string, error) {
	body := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.accessTokenURL,
		bytes.NewBuffer(bodyJSON),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %v", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	result := accessTokenResponse{}
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse access token response: %v", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("error fetching token: %v", result.Msg)
	}
	return result.TenantAccessToken, nil
}

func newAPI() *api {
	return &api{
		client:         http.DefaultClient,
		accessTokenURL: accessTokenURL,
	}
}
