package emailpassword

import (
	"context"
	"errors"
	"fmt"

	"github.com/etornam45/gorta/core"
)

type SignInResult struct {
	User    *core.User
	Session *core.Session
	Token   string
}

func (p *Plugin) SignIn(ctx context.Context, input SignInInput, meta core.SessionMeta) (*SignInResult, error) {
	input.Email = NormalizeEmail(input.Email)

	user, err := p.storage.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			core.VerifyPassword(input.Password, core.DUMMY_HASH) //nolint:errcheck
			return nil, core.ErrInvalidCredentials
		}
		fmt.Printf("[gorta] error finding user: %v\n", err)
		return nil, err
	}

	hash, err := p.creds.FindCredentialByUserID(ctx, user.ID)
	if err != nil {
		fmt.Printf("[gorta] error finding credential: %v\n", err)
		return nil, err
	}

	ok, err := core.VerifyPassword(input.Password, hash)
	if err != nil {
		return nil, fmt.Errorf("gorta: verifying password: %w", err)
	}
	if !ok {
		return nil, core.ErrInvalidCredentials
	}

	session, token, err := p.core.CreateSession(ctx, user.ID, meta)
	if err != nil {
		fmt.Printf("[gorta] error creating session: %v\n", err)
		return nil, err
	}

	return &SignInResult{
		User:    user,
		Session: session,
		Token:   token,
	}, nil
}
