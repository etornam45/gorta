package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

const (
	Google ProviderName = "google"
	GitHub ProviderName = "github"
	Yahoo  ProviderName = "yahoo"
	Apple  ProviderName = "apple"
)

type ProviderName string

type Provider interface {
	Name() string
	Config() *oauth2.Config
	GetUser(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}

type AuthOptionsProvider interface {
	AuthorizeOptions(state *OAuthState) ([]oauth2.AuthCodeOption, error)
	ExchangeOptions(state *OAuthState) []oauth2.AuthCodeOption
}

type UserInfo struct {
	ID            string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Raw           map[string]any
}
