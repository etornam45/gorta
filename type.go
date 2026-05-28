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
	Email         *string `json:"email"`
	Name          *string `json:"name"`
	Image         *string `json:"image"`
	EmailVerified *bool   `json:"email_verified"`
}

type SignUpInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type SignUpResult struct {
	User    *User
	Session *Session
	Token   string
}

type SignInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResult struct {
	User    *User
	Session *Session
	Token   string
}

type SessionMeta struct {
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}
