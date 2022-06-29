package auth

import (
	"os"
	"testing"

	"github.com/FreeLeh/GoFreeLeh/internal/google/fixtures"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

// Note that it is not really possible to test the "without stored credentials" path as
// it requires a real user input to get the auth code and also a mock HTTP server.
func TestNewOAuth2_CheckWrappedTransport_WithStoredCredentials(t *testing.T) {
	secretPath := fixtures.PathToFixture("client_secret.json")
	credsPath := fixtures.PathToFixture("stored_credentials.json")

	auth, err := NewOAuth2(secretPath, credsPath, []string{}, OAuth2Config{})
	assert.Nil(t, err, "should not have any error instantiating the OAuth2 wrapper")

	_, ok := auth.googleAuthClient.Transport.(*oauth2.Transport)
	assert.True(t, ok, "the HTTP client should be using the custom Google OAuth2 HTTP transport")

	_, err = os.Stat(credsPath)
	assert.Nil(t, err, "credential file should be created with the token info inside")
}
