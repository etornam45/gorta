package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/etornam45/gorta/internals"
)

func (a *Auth) SignUp(ctx context.Context, input internals.SignUpInput, meta internals.SessionMeta) (*internals.SignUpResult, error) {
	input.Email = normalizeEmail(input.Email)
	input.Name = sanitizeName(input.Name)
	fmt.Println("[gorta] input.Email: %s", input.Email)
	if err := validateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	_, err := a.adapter.FindUserByEmail(ctx, input.Email)
	if err == nil {
		return nil, internals.ErrUserAlreadyExists
	}
	if !errors.Is(err, internals.ErrUserNotFound) {
		a.logError("checking existing user", err)
		return nil, err
	}

	hash, err := internals.HashPassword(input.Password)
	if err != nil {
		a.logError("hashing password", err)
		return nil, err
	}

	if a.config.VerifyEmail {
		token, err := internals.GenerateToken()
		if err != nil {
			a.logError("generating token", err)
			return nil, err
		}
		verification, err := a.adapter.CreateVerification(ctx, internals.Verification{
			Identifier: input.Email,
			Token:      token,
			ExpiresAt:  time.Now().Add(time.Hour * 24 * 3),
		})
		if err != nil {
			a.logError("creating verification", err)
			return nil, fmt.Errorf("gorta: creating verification: %w", err)
		}
		if err := a.mailer.SendVerificationEmail(ctx, input.Email, fmt.Sprintf("http://%s/auth/verify-email?token=%s", a.config.VerifyEmailDomain, verification.Token)); err != nil {
			if err != nil {
				a.logError("sending verification email", err)
				return nil, fmt.Errorf("gorta: sending verification email: %w", err)
			}
		}
	}

	now := time.Now()
	user := internals.User{
		ID:        internals.GenerateID(),
		Email:     input.Email,
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdUser, err := a.adapter.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, internals.ErrUserAlreadyExists) {
			return nil, internals.ErrUserAlreadyExists
		}
		a.logError("creating user", err)
		return nil, err
	}

	if err := a.adapter.CreateCredential(ctx, createdUser.ID, hash); err != nil {
		if delErr := a.adapter.DeleteUser(ctx, createdUser.ID); delErr != nil {
			a.logError("cleaning up user after credential creation failure", delErr)
		}
		// a.logError("storing credential", err)
		return nil, fmt.Errorf("gorta: storing credential: %w", err)
	}

	session, token, err := a.createSession(ctx, createdUser.ID, meta)
	if err != nil {
		a.logError("creating session after sign-up", err)
		return &internals.SignUpResult{User: createdUser}, nil
	}

	return &internals.SignUpResult{
		User:    createdUser,
		Session: session,
		Token:   token,
	}, nil
}
