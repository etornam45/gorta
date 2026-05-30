package oauth

import (
	"context"
	"errors"
	"time"
)

var ErrOAuthStateNotFound = errors.New("oauth state not found")

type Storage interface {
	CreateAccount(ctx context.Context, account Account) (*Account, error)
	FindAccountByProvider(ctx context.Context, provider, providerAccountID string) (*Account, error)
}

// AuthState stores short-lived state tokens between authorize and callback.
// Expires in ~10 minutes. Use NewMemoryStateStore for single-server setups.
type AuthState interface {
	Save(ctx context.Context, state OAuthState) error
	Find(ctx context.Context, token string) (*OAuthState, error)
	Delete(ctx context.Context, token string) error
}

type Account struct {
	ID                string
	UserID            string
	Provider          ProviderName
	ProviderAccountID string
	AccessToken       string
	RefreshToken      string
	ExpiresAt         *time.Time
	CreatedAt         time.Time
}

type OAuthState struct {
	Token        string
	Provider     ProviderName
	ExpiresAt    time.Time
	CodeVerifier string
}
