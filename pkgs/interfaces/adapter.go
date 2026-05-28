package interfaces

import (
	"context"
	"github.com/etornam45/gorta/internals"
)

type Adapter interface {
	// Users
	CreateUser(ctx context.Context, user internals.User) (*internals.User, error)
	FindUserByID(ctx context.Context, id string) (*internals.User, error)
	FindUserByEmail(ctx context.Context, email string) (*internals.User, error)
	UpdateUser(ctx context.Context, id string, data internals.UpdateUserInput) (*internals.User, error)
	DeleteUser(ctx context.Context, id string) error

	// Credentials (email+password)
	CreateCredential(ctx context.Context, userID, passwordHash string) error
	FindCredentialByUserID(ctx context.Context, userID string) (passwordHash string, err error)
	UpdateCredential(ctx context.Context, userID, newHash string) error

	// Sessions
	CreateSession(ctx context.Context, session internals.Session) (*internals.Session, error)
	FindSessionByToken(ctx context.Context, token string) (*internals.Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionsByUserID(ctx context.Context, userID string) error

	// OAuth accounts
	CreateAccount(ctx context.Context, account internals.Account) (*internals.Account, error)
	FindAccountByProvider(ctx context.Context, provider, providerAccountID string) (*internals.Account, error)

	// Verifications (magic links, email verify, password reset)
	CreateVerification(ctx context.Context, v internals.Verification) (*internals.Verification, error)
	FindVerificationByToken(ctx context.Context, token string) (*internals.Verification, error)
	DeleteVerification(ctx context.Context, id string) error
}
