package magiclink

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
	CreateVerification(ctx context.Context, v Verification) (*Verification, error)
	FindVerificationByToken(ctx context.Context, token string) (*Verification, error)
	DeleteVerification(ctx context.Context, id string) error
}
