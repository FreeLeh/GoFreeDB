package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	refreshAccessTokenTimeout  = time.Second * 3
	refreshAccessTokenInterval = time.Second * 30
)

//go:generate mockgen -destination=api_mock.go -package=auth -build_constraint=gofreedb_test . accessTokenAPI
type accessTokenAPI interface {
	TenantAccessToken(
		ctx context.Context,
		appID string,
		appSecret string,
	) (string, error)
}

type TenantService struct {
	api       accessTokenAPI
	appID     string
	appSecret string

	tenantAccessToken     string
	tenantAccessTokenLock *sync.RWMutex

	logger    *slog.Logger
	config    Config
	done      chan struct{}
	closeOnce *sync.Once
}

func (s *TenantService) AccessToken() (string, error) {
	s.tenantAccessTokenLock.RLock()
	defer s.tenantAccessTokenLock.RUnlock()

	if s.tenantAccessToken == "" {
		return "", errors.New("tenant service returns empty access token")
	}
	return s.tenantAccessToken, nil
}

func (s *TenantService) runRefresher() {
	interval := s.config.RefreshInterval
	if interval == 0 {
		interval = refreshAccessTokenInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), refreshAccessTokenTimeout)
			if err := s.refresh(ctx); err != nil {
				slog.Error("error refreshing access token", "err", err)
			}
			cancel()
		}
	}
}

func (s *TenantService) refresh(ctx context.Context) error {
	t, err := s.api.TenantAccessToken(ctx, s.appID, s.appSecret)
	if err != nil {
		return err
	}

	s.tenantAccessTokenLock.Lock()
	defer s.tenantAccessTokenLock.Unlock()

	s.tenantAccessToken = t
	return nil
}

func (s *TenantService) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
	})
	return nil
}

func NewTenantService(
	appID string,
	appSecret string,
	config Config,
) (*TenantService, error) {
	s := TenantService{
		api:                   newAPI(),
		appID:                 appID,
		appSecret:             appSecret,
		tenantAccessToken:     "",
		tenantAccessTokenLock: &sync.RWMutex{},
		logger: slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{AddSource: true},
		)),
		config:    config,
		done:      make(chan struct{}, 1),
		closeOnce: &sync.Once{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.refresh(ctx); err != nil {
		return nil, err
	}

	go s.runRefresher()
	return &s, nil
}
