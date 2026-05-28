package gorta

import "context"

type Adapter interface {
	// Users
	CreateUser(ctx context.Context, user User) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, id string, data UpdateUserInput) (*User, error)
	DeleteUser(ctx context.Context, id string) error

	// Credentials (email+password)
	CreateCredential(ctx context.Context, userID, passwordHash string) error
	FindCredentialByUserID(ctx context.Context, userID string) (passwordHash string, err error)
	UpdateCredential(ctx context.Context, userID, newHash string) error

	// Sessions
	CreateSession(ctx context.Context, session Session) (*Session, error)
	FindSessionByToken(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionsByUserID(ctx context.Context, userID string) error

	// OAuth accounts
	CreateAccount(ctx context.Context, account Account) (*Account, error)
	FindAccountByProvider(ctx context.Context, provider, providerAccountID string) (*Account, error)

	// Verifications (magic links, email verify, password reset)
	CreateVerification(ctx context.Context, v Verification) (*Verification, error)
	FindVerificationByToken(ctx context.Context, token string) (*Verification, error)
	DeleteVerification(ctx context.Context, id string) error
}
