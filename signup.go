package gorta

import (
	"context"
	"errors"
	"fmt"
	"time"

)

func (a *Auth) SignUp(ctx context.Context, input SignUpInput, meta SessionMeta) (*SignUpResult, error) {
	input.Email = normalizeEmail(input.Email)
	input.Name = sanitizeName(input.Name)

	if err := validateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	_, err := a.adapter.FindUserByEmail(ctx, input.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, ErrUserNotFound) {
		a.logError("checking existing user", err)
		return nil, err
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		a.logError("hashing password", err)
		return nil, err
	}

	now := time.Now()
	user := User{
		ID:        generateID(),
		Email:     input.Email,
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdUser, err := a.adapter.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
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
		return &SignUpResult{User: createdUser}, nil
	}

	return &SignUpResult{
		User:    createdUser,
		Session: session,
		Token:   token,
	}, nil
}
