package auth

import (
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
)

// ServiceConfig defines a list of configurations that can be used to customise how the Google
// service account authentication flow works.
type ServiceConfig struct {
	// HTTPClient allows the client to customise the HTTP client used to perform the REST API calls.
	// This will be useful if you want to have a more granular control over the HTTP client (e.g. using a connection pool).
	HTTPClient *http.Client
}

// Service takes in service account relevant information and sets up *http.Client that can be used to access
// Google APIs seamlessly. Authentications will be handled automatically, including refreshing the access token
// when necessary.
type Service struct {
	googleAuthClient *http.Client
}

// HTTPClient returns a Google OAuth2 authenticated *http.Client that can be used to access Google APIs.
func (s *Service) HTTPClient() *http.Client {
	return s.googleAuthClient
}

// NewServiceFromFile creates a Service instance by reading the Google service account related information from a file.
//
// The "filePath" is referring to the service account JSON file that can be obtained by
// creating a new service account credentials in https://developers.google.com/identity/protocols/oauth2/service-account#creatinganaccount.
//
// The "scopes" tells Google what your application can do to your spreadsheets.
func NewServiceFromFile(filePath string, scopes Scopes, config ServiceConfig) (*Service, error) {
	authConfig, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return NewServiceFromJSON(authConfig, scopes, config)
}

// NewServiceFromJSON works exactly the same as NewServiceFromFile, but instead of reading from a file, the raw content
// of the Google service account JSON file is provided directly.
func NewServiceFromJSON(raw []byte, scopes Scopes, config ServiceConfig) (*Service, error) {
	c, err := google.JWTConfigFromJSON(raw, scopes...)
	if err != nil {
		return nil, err
	}

	return &Service{
		googleAuthClient: c.Client(getClientCtx(config.HTTPClient)),
	}, nil
}
