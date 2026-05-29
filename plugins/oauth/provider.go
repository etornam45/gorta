package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

const (
	Google ProviderName = "google"
	GitHub ProviderName = "github"
)

type ProviderName string

type Provider interface {
	Name() string
	Config() *oauth2.Config
	GetUser(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}

type UserInfo struct {
	ID            string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Raw           map[string]any
}
