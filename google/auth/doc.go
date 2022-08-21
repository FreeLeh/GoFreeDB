// Package auth provides general Google authentication implementation agnostic to what specific Google services or
// resources are used. Implementations in this package generate a https://pkg.go.dev/net/http#Client that can be used
// to access Google REST APIs seamlessly. Authentications will be handled automatically, including refreshing
//the access token when necessary.
package auth
