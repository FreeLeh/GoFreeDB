package auth

import (
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2/google"
)

type ServiceConfig struct {
	HTTPClient *http.Client
}

type Service struct {
	googleAuthClient *http.Client
}

func (s *Service) HTTPClient() *http.Client {
	return s.googleAuthClient
}

// TODO(EDWIN): need to document the fact that the service account is just an account.
// The client must make sure that the account has access to the Google Sheet.
// We may want to consider to have a hint printed in case ths this happens.
func NewService(filePath string, scopes Scopes, config ServiceConfig) (*Service, error) {
	authConfig, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	c, err := google.JWTConfigFromJSON(authConfig, scopes...)
	if err != nil {
		return nil, err
	}

	return &Service{
		googleAuthClient: c.Client(getClientCtx(config.HTTPClient)),
	}, nil
}
