package sqladapter

import (
	"context"
	"database/sql"
	"errors"

	"github.com/etornam45/gorta"
)

type SQLAdapter struct {
	db *sql.DB
}

func (a *SQLAdapter) CreateUser(ctx context.Context, user gorta.User) (*gorta.User, error) {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO users (id, email, name, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
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
	return &u, err
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
	return &u, err
}

func (a *SQLAdapter) UpdateUser(ctx context.Context, id string, data gorta.UpdateUserInput) (*gorta.User, error) {
	var u gorta.User
	err := a.db.QueryRowContext(ctx,
		`UPDATE users SET email = $1, email_verified = $2, name = $3, image = $4
		 WHERE id = $5
		 RETURNING id, email, email_verified, name, image, created_at, updated_at`,
		data.Email, data.EmailVerified, data.Name, data.Image, id,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.Name, &u.Image, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorta.ErrUserNotFound
	}
	return &u, err
}

func (a *SQLAdapter) DeleteUser(ctx context.Context, id string) error {
	_, err := a.db.ExecContext(ctx,
		`DELETE FROM users WHERE id = $1`, id,
	)
	return err
}