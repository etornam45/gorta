package gorta

import (
	"context"
	"time"
)

func (a *Auth) VerifyEmail(ctx context.Context, token string) (*User, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	verification, err := a.adapter.FindVerificationByToken(ctx, token)
	if err != nil {
		a.logError("finding verification", err)
		return nil, ErrVerificationNotFound
	}
	if verification.ExpiresAt.Before(time.Now()) {
		a.logError("verification expired", err)
		return nil, ErrVerificationExpired
	}
	user, err := a.adapter.FindUserByEmail(ctx, verification.Identifier)
	if err != nil {
		a.logError("finding user", err)
		return nil, ErrUserNotFound
	}
	if user.EmailVerified {
		return nil, ErrEmailNotVerified
	}
	_, err = a.adapter.UpdateUser(ctx, user.ID, UpdateUserInput{
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