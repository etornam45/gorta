package emailpassword

import (
	"context"
	"time"
)

type Verification struct {
	ID         string
	Identifier string // the email address
	Token      string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

type Storage interface {
	EmailPasswordFindVerificationByToken(ctx context.Context, token string) (*Verification, error)
	DeleteVerification(ctx context.Context, id string) error
	CreateCredential(ctx context.Context, userID, passwordHash string) error
	FindCredentialByUserID(ctx context.Context, userID string) (passwordHash string, err error)
	UpdateCredential(ctx context.Context, userID, newHash string) error
}
