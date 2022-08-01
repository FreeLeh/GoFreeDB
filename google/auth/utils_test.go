package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestGetClientCtx(t *testing.T) {
	t.Run("without_custom_client", func(t *testing.T) {
		ctx := getClientCtx(nil)
		client := ctx.Value(oauth2.HTTPClient)
		assert.Nil(t, client, "client should be nil")
	})

	t.Run("with_custom_client", func(t *testing.T) {
		custom := &http.Client{}
		ctx := getClientCtx(custom)
		client := ctx.Value(oauth2.HTTPClient).(*http.Client)

		assert.NotNil(t, client, "client should not be nil")
		assert.True(t, custom == client, "client should be the given custom HTTP client")
	})
}
