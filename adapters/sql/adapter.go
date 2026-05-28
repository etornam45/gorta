package sqladapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/etornam45/gorta"
)

type SQLAdapter struct {
	db *sql.DB
}

func New(db *sql.DB) gorta.Adapter {
	return &SQLAdapter{db: db}
}


func (a *SQLAdapter) CreateUser(ctx context.Context, user gorta.User) (*gorta.User, error) {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO users (id, email, email_verified, name, image, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Email, user.EmailVerified, user.Name, user.Image,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, gorta.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("sqladapter: creating user: %w", err)
	}
	return &user, nil
}

func (a *SQLAdapter) FindUserByID(ctx context.Context, id string) (*gorta.User, error) {
	var u gorta.User
	err := a.db.QueryRowContext(ctx,
		`SELECT id, email, email_verified, name, image, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.Name, &u.Image, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorta.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("sqladapter: finding user by id: %w", err)
	}
	return &u, nil
}

func (a *SQLAdapter) FindUserByEmail(ctx context.Context, email string) (*gorta.User, error) {
	var u gorta.User
	err := a.db.QueryRowContext(ctx,
		`SELECT id, email, email_verified, name, image, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.Name, &u.Image, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorta.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("sqladapter: finding user by email: %w", err)
	}
	return &u, nil
}

func (a *SQLAdapter) UpdateUser(ctx context.Context, id string, input gorta.UpdateUserInput) (*gorta.User, error) {
	if input.Name != nil {
		if _, err := a.db.ExecContext(ctx,
			`UPDATE users SET name = $1, updated_at = $2 WHERE id = $3`,
			*input.Name, time.Now(), id,
		); err != nil {
			return nil, fmt.Errorf("sqladapter: updating user name: %w", err)
		}
	}
	if input.Image != nil {
		if _, err := a.db.ExecContext(ctx,
			`UPDATE users SET image = $1, updated_at = $2 WHERE id = $3`,
			*input.Image, time.Now(), id,
		); err != nil {
			return nil, fmt.Errorf("sqladapter: updating user image: %w", err)
		}
	}
	if input.EmailVerified != nil {
		if _, err := a.db.ExecContext(ctx,
			`UPDATE users SET email_verified = $1, updated_at = $2 WHERE id = $3`,
			*input.EmailVerified, time.Now(), id,
		); err != nil {
			return nil, fmt.Errorf("sqladapter: updating email_verified: %w", err)
		}
	}
	return a.FindUserByID(ctx, id)
}

func (a *SQLAdapter) DeleteUser(ctx context.Context, id string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (a *SQLAdapter) CreateCredential(ctx context.Context, userID, passwordHash string) error {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO credentials (id, user_id, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $4)`,
		newID(), userID, passwordHash, time.Now(),
	)
	return err
}

func (a *SQLAdapter) FindCredentialByUserID(ctx context.Context, userID string) (string, error) {
	var hash string
	err := a.db.QueryRowContext(ctx,
		`SELECT password_hash FROM credentials WHERE user_id = $1`, userID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", gorta.ErrUserNotFound
	}
	return hash, err
}

func (a *SQLAdapter) UpdateCredential(ctx context.Context, userID, newHash string) error {
	_, err := a.db.ExecContext(ctx,
		`UPDATE credentials SET password_hash = $1, updated_at = $2 WHERE user_id = $3`,
		newHash, time.Now(), userID,
	)
	return err
}

func (a *SQLAdapter) CreateSession(ctx context.Context, s gorta.Session) (*gorta.Session, error) {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, token, expires_at, ip_address, user_agent, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		s.ID, s.UserID, s.Token, s.ExpiresAt, s.IPAddress, s.UserAgent, s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("sqladapter: creating session: %w", err)
	}
	return &s, nil
}

func (a *SQLAdapter) FindSessionByToken(ctx context.Context, token string) (*gorta.Session, error) {
	var s gorta.Session
	err := a.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, expires_at, ip_address, user_agent, created_at
		 FROM sessions WHERE token = $1`, token,
	).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.IPAddress, &s.UserAgent, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorta.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("sqladapter: finding session: %w", err)
	}
	return &s, nil
}

func (a *SQLAdapter) DeleteSession(ctx context.Context, id string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (a *SQLAdapter) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (a *SQLAdapter) CreateAccount(ctx context.Context, acc gorta.Account) (*gorta.Account, error) {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		acc.ID, acc.UserID, acc.Provider, acc.ProviderAccountID,
		acc.AccessToken, acc.RefreshToken, acc.ExpiresAt, acc.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("sqladapter: creating account: %w", err)
	}
	return &acc, nil
}

func (a *SQLAdapter) FindAccountByProvider(ctx context.Context, provider, providerAccountID string) (*gorta.Account, error) {
	var acc gorta.Account
	err := a.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, created_at
		 FROM accounts WHERE provider = $1 AND provider_account_id = $2`,
		provider, providerAccountID,
	).Scan(&acc.ID, &acc.UserID, &acc.Provider, &acc.ProviderAccountID,
		&acc.AccessToken, &acc.RefreshToken, &acc.ExpiresAt, &acc.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorta.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("sqladapter: finding account: %w", err)
	}
	return &acc, nil
}

func (a *SQLAdapter) CreateVerification(ctx context.Context, v gorta.Verification) (*gorta.Verification, error) {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO verifications (id, identifier, token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		v.ID, v.Identifier, v.Token, v.ExpiresAt, v.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("sqladapter: creating verification: %w", err)
	}
	return &v, nil
}

func (a *SQLAdapter) FindVerificationByToken(ctx context.Context, token string) (*gorta.Verification, error) {
	var v gorta.Verification
	err := a.db.QueryRowContext(ctx,
		`SELECT id, identifier, token, expires_at, created_at
		 FROM verifications WHERE token = $1`, token,
	).Scan(&v.ID, &v.Identifier, &v.Token, &v.ExpiresAt, &v.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorta.ErrVerificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("sqladapter: finding verification: %w", err)
	}
	return &v, nil
}

func (a *SQLAdapter) DeleteVerification(ctx context.Context, id string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM verifications WHERE id = $1`, id)
	return err
}

func isUniqueViolation(err error) bool {
	_ = err
	return false
}

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}