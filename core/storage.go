package core

import (
	"context"
	"time"
)

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Name          string    `json:"name"`
	Image         string    `json:"image,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	User      *User     `json:"user,omitempty"`
}

type UpdateUserInput struct {
	Name          *string
	Image         *string
	EmailVerified *bool
}

type SessionMeta struct {
	IPAddress string
	UserAgent string
}

type Storage interface {
	CreateUser(ctx context.Context, user User) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*User, error)
	DeleteUser(ctx context.Context, id string) error

	CreateSession(ctx context.Context, session Session) (*Session, error)
	FindSessionByToken(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteSessionsByUserID(ctx context.Context, userID string) error
	RevokeSession(ctx context.Context, token string) error
	RevokeAllSessions(ctx context.Context, userID string) error
}