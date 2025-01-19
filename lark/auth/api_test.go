package auth

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tt.statusCode)

				// For malformed JSON test case
				if tt.returnMalformedJSON {
					_, err := w.Write([]byte(`{malformed json`))
					assert.NoError(t, err)
					return
				}

				// Write response body for other cases
				if tt.serverResponse != nil {
					err := json.NewEncoder(w).Encode(tt.serverResponse)
					assert.NoError(t, err)
				}
			}))
			defer server.Close()

			// Need to replace the URL with the httptest server as well.
			a := &api{
				client:         server.Client(),
				accessTokenURL: server.URL,
			}

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
