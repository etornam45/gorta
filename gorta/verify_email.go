package gorta

import (
	"context"
	"time"

	"github.com/etornam45/gorta/internals"
)

func (a *Auth) VerifyEmail(ctx context.Context, token string) (*internals.User, error) {
	if token == "" {
		return nil, internals.ErrInvalidToken
	}
	verification, err := a.adapter.FindVerificationByToken(ctx, token)
	if err != nil {
		a.logError("finding verification", err)
		return nil, internals.ErrVerificationNotFound
	}
	if verification.ExpiresAt.Before(time.Now()) {
		a.logError("verification expired", err)
		return nil, internals.ErrVerificationExpired
	}
	user, err := a.adapter.FindUserByEmail(ctx, verification.Identifier)
	if err != nil {
		a.logError("finding user", err)
		return nil, internals.ErrUserNotFound
	}
	if user.EmailVerified {
		return nil, internals.ErrEmailNotVerified
	}
	_, err = a.adapter.UpdateUser(ctx, user.ID, internals.UpdateUserInput{
		EmailVerified: &[]bool{true}[0],
	})
	if err != nil {
		a.logError("updating user", err)
		return nil, err
	}

	// clean up the verification
	if err := a.adapter.DeleteVerification(ctx, verification.ID); err != nil {
		a.logError("deleting verification", err)
	}
	return user, nil
}
