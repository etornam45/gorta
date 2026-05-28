package gorta

import "time"

type User struct {
	ID            string
	Email         string
	EmailVerified bool
	Name          string
	Image         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	CreatedAt time.Time
	User      *User
}

type Account struct {
	ID                string
	UserID            string
	Provider          string
	ProviderAccountID string
	AccessToken       string
	RefreshToken      string
	ExpiresAt         *time.Time
	CreatedAt         time.Time
}


type Verification struct {
	ID         string
	Identifier string // the email address
	Token      string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}


type UpdateUserInput struct {
	Email         *string
	Name          *string
	Image         *string
	EmailVerified *bool
}

type SignUpInput struct {
	Email    string
	Password string
	Name     string
}

type SignUpResult struct {
	User    *User
	Session *Session
	Token   string
}

type SignInInput struct {
	Email    string
	Password string
}

type SignInResult struct {
	User    *User
	Session *Session
	Token   string
}

type SessionMeta struct {
	IPAddress string
	UserAgent string
}