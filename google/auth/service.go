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

func NewServiceFromFile(filePath string, scopes Scopes, config ServiceConfig) (*Service, error) {
	authConfig, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return NewServiceFromJSON(authConfig, scopes, config)
}

func NewServiceFromJSON(raw []byte, scopes Scopes, config ServiceConfig) (*Service, error) {
	c, err := google.JWTConfigFromJSON(raw, scopes...)
	if err != nil {
		return nil, err
	}

	return &Service{
		googleAuthClient: c.Client(getClientCtx(config.HTTPClient)),
	}, nil
}
