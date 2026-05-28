package core

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrAccountNotFound      = errors.New("account not found")
	ErrVerificationNotFound = errors.New("verification not found")
	ErrVerificationExpired  = errors.New("verification expired")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidToken         = errors.New("invalid token")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrWeakPassword         = errors.New("password must be at least 8 characters")
	ErrInvalidEmail         = errors.New("invalid email address")
)

type AuthError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
}

func (e *AuthError) Error() string { return e.Code + ": " + e.Message }
func (e *AuthError) Unwrap() error { return e.Err }

func newAuthError(code, message string, statusCode int, err error) *AuthError {
	return &AuthError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}
