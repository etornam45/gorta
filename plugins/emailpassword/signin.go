package emailpassword

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/etornam45/gorta/core"
	"github.com/etornam45/gorta/internals"
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
		if errors.Is(err, internals.ErrUserNotFound) {
			internals.VerifyPassword(input.Password, internals.DUMMY_HASH) //nolint:errcheck
			return nil, internals.ErrInvalidCredentials
		}
		fmt.Printf("[gorta] error finding user: %v\n", err)
		return nil, err
	}

	hash, err := p.creds.FindCredentialByUserID(ctx, user.ID)
	if err != nil {
		fmt.Printf("[gorta] error finding credential: %v\n", err)
		return nil, err
	}

	ok, err := internals.VerifyPassword(input.Password, hash)
	if err != nil {
		return nil, fmt.Errorf("gorta: verifying password: %w", err)
	}
	if !ok {
		return nil, internals.ErrInvalidCredentials
	}
	storage, ok := p.storage.(core.Storage)
	if !ok {
		fmt.Printf("[gorta] error creating session: %v\n", err)
		return nil, err
	}
	token, err := core.GenerateToken()
	if err != nil {
		fmt.Printf("[gorta] error generating token: %v\n", err)
		return nil, err
	}
	session, err := storage.CreateSession(ctx, core.Session{
		ID:        core.GenerateID(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // TODO: use configurations for session duration
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
	})
	if err != nil {
		fmt.Printf("[gorta] error verifying password: %v\n", err)
		return nil, err
	}

	return &SignInResult{
		User:    user,
		Session: session,
		Token:   token,
	}, nil
}
