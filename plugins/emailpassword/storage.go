package emailpassword

import "context"

type Storage interface {
	CreateCredential(ctx context.Context, userID, passwordHash string) error
	FindCredentialByUserID(ctx context.Context, userID string) (passwordHash string, err error)
	UpdateCredential(ctx context.Context, userID, newHash string) error
}
