package auth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

func getClientCtx(customClient *http.Client) context.Context {
	ctx := context.Background()
	if customClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, customClient)
	}
	return ctx
}
