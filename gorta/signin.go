package gorta

import (
	"context"
	"errors"
	"fmt"

	"github.com/etornam45/gorta/internals"
)

func (a *Auth) SignIn(ctx context.Context, input internals.SignInInput, meta internals.SessionMeta) (*internals.SignInResult, error) {
	input.Email = normalizeEmail(input.Email)

	user, err := a.adapter.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, internals.ErrUserNotFound) {
			internals.VerifyPassword(input.Password, internals.DUMMY_HASH) //nolint:errcheck
			return nil, internals.ErrInvalidCredentials
		}
		a.logError("finding user", err)
		return nil, err
	}

	hash, err := a.adapter.FindCredentialByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("gorta: finding credential: %w", err)
	}

	ok, err := internals.VerifyPassword(input.Password, hash)
	if err != nil {
		return nil, fmt.Errorf("gorta: verifying password: %w", err)
	}
	if !ok {
		return nil, internals.ErrInvalidCredentials
	}

	session, token, err := a.createSession(ctx, user.ID, meta)
	if err != nil {
		a.logError("verifying password", err)
		return nil, err
	}

	return &internals.SignInResult{
		User:    user,
		Session: session,
		Token:   token,
	}, nil
}
