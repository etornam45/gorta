package gorta

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (a *Auth) createSession(ctx context.Context, userID string, meta SessionMeta) (*Session, string, error) {
	token, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("gorta: generating session token: %w", err)
	}

	session := Session{
		ID:        generateID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(a.config.SessionDuration),
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: time.Now(),
	}

	created, err := a.adapter.CreateSession(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("gorta: creating session: %w", err)
	}

	return created, token, nil
}

func (a *Auth) ValidateSession(ctx context.Context, token string) (*Session, error) {
	session, err := a.adapter.FindSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		a.logError("finding session", err)
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		go func() {
			if delErr := a.adapter.DeleteSession(context.Background(), session.ID); delErr != nil {
				a.logError("deleting expired session", delErr)
			}
		}()
		return nil, ErrSessionExpired
	}

	user, err := a.adapter.FindUserByID(ctx, session.UserID)
	if err != nil {
		a.logError("finding session user", err)
		return nil, err
	}

	session.User = user
	return session, nil
}

func (a *Auth) RevokeSession(ctx context.Context, token string) error {
	session, err := a.adapter.FindSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil
		}
		a.logError("finding session to revoke", err)
		return err
	}

	if err := a.adapter.DeleteSession(ctx, session.ID); err != nil {
		a.logError("revoking session", err)
		return err
	}

	return nil
}

func (a *Auth) RevokeAllSessions(ctx context.Context, userID string) error {
	if err := a.adapter.DeleteSessionsByUserID(ctx, userID); err != nil {
		a.logError("revoking all sessions", err)
		return err
	}
	return nil
}
