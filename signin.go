package gorta

import (
	"context"
	"errors"
	"fmt"
)

func (a *Auth) SignIn(ctx context.Context, input SignInInput, meta SessionMeta) (*SignInResult, error) {
	input.Email = normalizeEmail(input.Email)

	user, err := a.adapter.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			verifyPassword(input.Password, dummyHash) //nolint:errcheck
			return nil, ErrInvalidCredentials
		}
		a.logError("finding user", err)
		return nil, err
	}

	hash, err := a.adapter.FindCredentialByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("gorta: finding credential: %w", err)
	}

	ok, err := verifyPassword(input.Password, hash)
	if err != nil {
		return nil, fmt.Errorf("gorta: verifying password: %w", err)
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	session, token, err := a.createSession(ctx, user.ID, meta)
	if err != nil {
		a.logError("verifying password", err)
		return nil, err
	}

	return &SignInResult{
		User:    user,
		Session: session,
		Token:   token,
	}, nil
}
