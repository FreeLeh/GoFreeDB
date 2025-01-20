package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// newTestTenantService creates a TenantService instance for testing
func newTestTenantService(api accessTokenAPI, appID, appSecret string, config Config) *TenantService {
	return &TenantService{
		api:                   api,
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
}

func TestTenantService_AccessToken(t *testing.T) {
	t.Run("returns error when token is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockAPI := NewMockaccessTokenAPI(ctrl)

		svc := newTestTenantService(
			mockAPI,
			"app-id",
			"app-secret",
			Config{},
		)
		svc.tenantAccessToken = "" // Force empty token for test

		token, err := svc.AccessToken()
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

func TestTenantService_refresh(t *testing.T) {
	t.Run("successfully refreshes token", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := NewMockaccessTokenAPI(ctrl)
		mockAPI.EXPECT().
			TenantAccessToken(gomock.Any(), "app-id", "app-secret").
			Return("new-token", nil)

		svc := newTestTenantService(
			mockAPI,
			"app-id",
			"app-secret",
			Config{},
		)

		err := svc.refresh(context.Background())
		assert.NoError(t, err)

		token, err := svc.AccessToken()
		assert.NoError(t, err)
		assert.Equal(t, "new-token", token)
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := NewMockaccessTokenAPI(ctrl)
		mockAPI.EXPECT().
			TenantAccessToken(gomock.Any(), "app-id", "app-secret").
			Return("", errors.New("api error"))

		svc := newTestTenantService(
			mockAPI,
			"app-id",
			"app-secret",
			Config{},
		)

		err := svc.refresh(context.Background())
		assert.Error(t, err)
	})
}

func TestTenantService_runRefresher(t *testing.T) {
	t.Run("refreshes token periodically", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := NewMockaccessTokenAPI(ctrl)
		gomock.InOrder(
			mockAPI.EXPECT().
				TenantAccessToken(gomock.Any(), "app-id", "app-secret").
				Return("token-1", nil).
				Times(1),
			mockAPI.EXPECT().
				TenantAccessToken(gomock.Any(), "app-id", "app-secret").
				Return("token-2", nil).
				AnyTimes(),
		)

		svc := newTestTenantService(
			mockAPI,
			"app-id",
			"app-secret",
			Config{
				RefreshInterval: time.Millisecond * 50,
			},
		)

		go svc.runRefresher()
		defer svc.Close()

		// Wait for first refresh
		time.Sleep(time.Millisecond * 75)
		token, err := svc.AccessToken()
		assert.NoError(t, err)
		assert.Equal(t, "token-1", token)

		// Wait for second refresh
		time.Sleep(time.Millisecond * 150)
		token, err = svc.AccessToken()
		assert.NoError(t, err)
		assert.Equal(t, "token-2", token)
	})

	t.Run("stops refreshing when closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := NewMockaccessTokenAPI(ctrl)
		mockAPI.EXPECT().
			TenantAccessToken(gomock.Any(), "app-id", "app-secret").
			Return("token", nil).
			Times(1) // Should only be called once before close

		svc := newTestTenantService(
			mockAPI,
			"app-id",
			"app-secret",
			Config{
				RefreshInterval: time.Millisecond * 50,
			},
		)

		go svc.runRefresher()

		// Wait for first refresh
		time.Sleep(time.Millisecond * 75)

		err := svc.Close()
		assert.NoError(t, err)

		// Wait longer than refresh interval to ensure no more refreshes
		time.Sleep(time.Millisecond * 100)
		token, err := svc.AccessToken()
		assert.NoError(t, err)
		assert.Equal(t, "token", token)
	})
}

func TestTenantService_Concurrency(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockAPI := NewMockaccessTokenAPI(ctrl)
	mockAPI.EXPECT().
		TenantAccessToken(gomock.Any(), "app-id", "app-secret").
		Return("concurrent-token", nil).
		AnyTimes()

	svc := newTestTenantService(
		mockAPI,
		"app-id",
		"app-secret",
		Config{RefreshInterval: time.Millisecond},
	)
	defer svc.Close()

	go svc.runRefresher()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 10)

			for j := 0; j < 100; j++ {
				token, err := svc.AccessToken()
				assert.NoError(t, err)
				assert.Equal(t, "concurrent-token", token)
			}
		}()
	}
	wg.Wait()
}
