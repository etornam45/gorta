package emailpassword

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/etornam45/gorta/core"
	"github.com/etornam45/gorta/internals"
	"github.com/etornam45/gorta/plugins/magiclink"
)

type SignUpResult struct {
	User    core.User
	Session *core.Session
	Token   string
}

func (p *Plugin) SignUp(ctx context.Context, input SignUpInput, meta core.SessionMeta) (*SignUpResult, error) {
	input.Email = NormalizeEmail(input.Email)
	input.Name = SanitizeName(input.Name)
	fmt.Println("[gorta] input.Email: %s", input.Email)
	if err := ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	_, err := p.storage.FindUserByEmail(ctx, input.Email)
	if err == nil {
		return nil, internals.ErrUserAlreadyExists
	}
	if !errors.Is(err, internals.ErrUserNotFound) {
		fmt.Printf("[gorta] error checking existing user: %v\n", err)
		return nil, err
	}

	hash, err := internals.HashPassword(input.Password)
	if err != nil {
		fmt.Printf("[gorta] error hashing password: %v\n", err)
		return nil, err
	}

	if p.config.VerifyEmail {
		token, err := internals.GenerateToken()
		if err != nil {
			fmt.Printf("[gorta] error generating token: %v\n", err)
			return nil, err
		}

		magiclinkStorage, ok := p.storage.(magiclink.Storage)
		if !ok {
			fmt.Printf("[gorta] error creating verification: %v\n", err)
			return nil, err
		}
		verification, err := magiclinkStorage.CreateVerification(ctx, magiclink.Verification{
			Identifier: input.Email,
			Token:      token,
			ExpiresAt:  time.Now().Add(time.Hour * 24 * 3),
		})
		if err != nil {
			fmt.Printf("[gorta] error creating verification: %v\n", err)
			return nil, fmt.Errorf("gorta: creating verification: %w", err)
		}

		mailer, ok := p.mailer.(Mailer)
		if !ok {
			fmt.Printf("[gorta] error sending verification email: %v\n", err)
			return nil, err
		}
		err = mailer.SendVerificationEmail(ctx, input.Email, fmt.Sprintf("http://%s/auth/verify-email?token=%s", p.config.VerifyEmailDomain, verification.Token))
		if err != nil {
			fmt.Printf("[gorta] error sending verification email: %v\n", err)
			return nil, err
		}
	}

	now := time.Now()
	user := core.User{
		ID:        core.GenerateID(),
		Email:     input.Email,
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdUser, err := p.storage.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, internals.ErrUserAlreadyExists) {
			return nil, internals.ErrUserAlreadyExists
		}
		fmt.Printf("[gorta] error creating user: %v\n", err)
		return nil, err
	}

	if err := p.creds.CreateCredential(ctx, createdUser.ID, hash); err != nil {
		if delErr := p.storage.DeleteUser(ctx, createdUser.ID); delErr != nil {
			fmt.Printf("[gorta] error cleaning up user after credential creation failure: %v\n", delErr)
		}
		fmt.Printf("[gorta] error storing credential: %v\n", err)
		return nil, fmt.Errorf("gorta: storing credential: %w", err)
	}

	token, err := core.GenerateToken()
	if err != nil {
		fmt.Printf("[gorta] error generating token: %v\n", err)
		return nil, err
	}

	session, err := p.storage.CreateSession(ctx, core.Session{
		ID:        core.GenerateID(),
		UserID:    createdUser.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
	})
	if err != nil {
		fmt.Printf("[gorta] error creating session after sign-up: %v\n", err)
		return nil, err
	}

	return &SignUpResult{
		User:    *createdUser,
		Session: session,
		Token:   token,
	}, nil
}