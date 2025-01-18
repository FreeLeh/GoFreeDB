package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	stateLength  = 32
)

// OAuth2Config defines a list of configurations that can be used to customise how the Google OAuth2 flow works.
type OAuth2Config struct {
	// HTTPClient allows the client to customise the HTTP client used to perform the REST API calls.
	// This will be useful if you want to have a more granular control over the HTTP client (e.g. using a connection pool).
	HTTPClient *http.Client
}

// OAuth2 takes in OAuth2 relevant information and sets up *http.Client that can be used to access
// Google APIs seamlessly. Authentications will be handled automatically, including refreshing the access token
// when necessary.
type OAuth2 struct {
	googleAuthClient *http.Client
}

// HTTPClient returns a Google OAuth2 authenticated *http.Client that can be used to access Google APIs.
func (o *OAuth2) HTTPClient() *http.Client {
	return o.googleAuthClient
}

// NewOAuth2FromFile creates an OAuth2 instance by reading the OAuth2 related information from a secret file.
//
// The "secretFilePath" is referring to the OAuth2 credentials JSON file that can be obtained by
// creating a new OAuth2 credentials in https://console.cloud.google.com/apis/credentials.
// You can put any link for the redirection URL field.
//
// The "credsFilePath" is referring to a file where the generated access and refresh token will be cached.
// This file will be created automatically once the OAuth2 authentication is successful.
//
// The "scopes" tells Google what your application can do to your spreadsheets.
//
// Note that since this is an OAuth2 server flow, human interaction will be needed for the very first authentication.
// During the OAuth2 flow, you will be asked to click a generated URL in the terminal.
func NewOAuth2FromFile(secretFilePath string, credsFilePath string, scopes Scopes, config OAuth2Config) (*OAuth2, error) {
	rawAuthConfig, err := os.ReadFile(secretFilePath)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(credsFilePath); err != nil {
		return newFromClientSecret(rawAuthConfig, credsFilePath, scopes, config)
	}

	rawCreds, err := os.ReadFile(credsFilePath)
	if err != nil {
		return nil, err
	}
	return newFromStoredCreds(rawAuthConfig, rawCreds, scopes, config)
}

func newFromStoredCreds(rawAuthConfig []byte, rawCreds []byte, scopes Scopes, config OAuth2Config) (*OAuth2, error) {
	var token oauth2.Token
	if err := json.Unmarshal(rawCreds, &token); err != nil {
		return nil, err
	}

	c, err := google.ConfigFromJSON(rawAuthConfig, scopes...)
	if err != nil {
		return nil, err
	}

	return &OAuth2{
		googleAuthClient: c.Client(getClientCtx(config.HTTPClient), &token),
	}, nil
}

func newFromClientSecret(rawAuthConfig []byte, credsFilePath string, scopes Scopes, config OAuth2Config) (*OAuth2, error) {
	c, err := google.ConfigFromJSON(rawAuthConfig, scopes...)
	if err != nil {
		return nil, err
	}

	authCode, err := getAuthCode(c)
	if err != nil {
		return nil, err
	}

	token, err := getToken(c, authCode)
	if err != nil {
		return nil, err
	}
	if err := storeCredentials(credsFilePath, token); err != nil {
		return nil, err
	}

	return &OAuth2{
		googleAuthClient: c.Client(getClientCtx(config.HTTPClient), token),
	}, nil
}

func getAuthCode(c *oauth2.Config) (string, error) {
	state := generateState()
	authCodeURL := c.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Printf("Visit the URL for the auth dialog: %v\n", authCodeURL)
	fmt.Print("Paste the redirection URL here: ")

	var rawRedirectionURL string
	if _, err := fmt.Scan(&rawRedirectionURL); err != nil {
		return "", err
	}

	redirectionURL, err := url.Parse(rawRedirectionURL)
	if err != nil {
		return "", err
	}

	query := redirectionURL.Query()
	if query.Get("state") != state {
		return "", errors.New("oauth state does not match")
	}
	return query.Get("code"), nil
}

func generateState() string {
	sb := strings.Builder{}
	randSrc := rand.NewSource(time.Now().UnixMilli())

	for i := 0; i < stateLength; i++ {
		idx := randSrc.Int63() % int64(len(alphanumeric))
		sb.WriteByte(alphanumeric[idx])
	}

	return sb.String()
}

func getToken(c *oauth2.Config, authCode string) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return c.Exchange(ctx, authCode)
}

func storeCredentials(credsFilePath string, token *oauth2.Token) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(credsFilePath, tokenJSON, 0644)
}
