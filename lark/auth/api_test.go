package auth

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	"github.com/h2non/gock"
)

func TestTenantAccessToken(t *testing.T) {
	tests := []struct {
		name                string
		serverResponse      *accessTokenResponse
		returnMalformedJSON bool
		statusCode          int
		appID               string
		appSecret           string
		expectError         bool
		expectedToken       string
	}{
		{
			name: "successful response",
			serverResponse: &accessTokenResponse{
				Code:              0,
				Msg:               "ok",
				TenantAccessToken: "valid-token-123",
			},
			returnMalformedJSON: false,
			statusCode:          http.StatusOK,
			appID:               "test-app-id",
			appSecret:           "test-app-secret",
			expectError:         false,
			expectedToken:       "valid-token-123",
		},
		{
			name: "error response from server",
			serverResponse: &accessTokenResponse{
				Code: 10002,
				Msg:  "invalid app_id",
			},
			returnMalformedJSON: false,
			statusCode:          http.StatusOK,
			appID:               "invalid-app-id",
			appSecret:           "test-app-secret",
			expectError:         true,
		},
		{
			name:                "server returns non-200 status",
			returnMalformedJSON: false,
			statusCode:          http.StatusInternalServerError,
			appID:               "test-app-id",
			appSecret:           "test-app-secret",
			expectError:         true,
		},
		{
			name:                "server returns malformed JSON",
			returnMalformedJSON: true,
			statusCode:          http.StatusOK,
			appID:               "test-app-id",
			appSecret:           "test-app-secret",
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()

			resp := gock.New(accessTokenURL).
				Post("").
				MatchHeader("Content-Type", "application/json").
				Reply(tt.statusCode)

			if tt.returnMalformedJSON {
				resp.BodyString(`{malformed json`)
			} else if tt.serverResponse != nil {
				resp.JSON(tt.serverResponse)
			}

			a := newAPI()

			token, err := a.TenantAccessToken(context.Background(), tt.appID, tt.appSecret)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}
