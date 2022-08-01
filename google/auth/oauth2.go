package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

type OAuth2Config struct {
	HTTPClient *http.Client
}

type OAuth2 struct {
	googleAuthClient *http.Client
}

func (o *OAuth2) HTTPClient() *http.Client {
	return o.googleAuthClient
}

func NewOAuth2FromFile(secretFilePath string, credsFilePath string, scopes Scopes, config OAuth2Config) (*OAuth2, error) {
	rawAuthConfig, err := ioutil.ReadFile(secretFilePath)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(credsFilePath); err != nil {
		return newFromClientSecret(rawAuthConfig, credsFilePath, scopes, config)
	}

	rawCreds, err := ioutil.ReadFile(credsFilePath)
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
	return ioutil.WriteFile(credsFilePath, tokenJSON, 0644)
}
