package core

import (
	"encoding/json"
	"errors"
	"net/http"

)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return
	}
}

func WriteError(w http.ResponseWriter, err error) {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		WriteJSON(w, authErr.StatusCode, map[string]string{
			"error":   authErr.Code,
			"message": authErr.Message,
		})
		return
	}

	status, code, message := sentinelToHTTP(err)
	WriteJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func sentinelToHTTP(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password"
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound, "USER_NOT_FOUND", "user not found"
	case errors.Is(err, ErrUserAlreadyExists):
		return http.StatusConflict, "USER_ALREADY_EXISTS", "a user with this email already exists"
	case errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrInvalidToken):
		return http.StatusUnauthorized, "INVALID_SESSION", "session is invalid"
	case errors.Is(err, ErrSessionExpired):
		return http.StatusUnauthorized, "SESSION_EXPIRED", "session has expired"
	case errors.Is(err, ErrEmailNotVerified):
		return http.StatusForbidden, "EMAIL_NOT_VERIFIED", "please verify your email address"
	case errors.Is(err, ErrWeakPassword):
		return http.StatusBadRequest, "WEAK_PASSWORD", "password must be at least 8 characters"
	case errors.Is(err, ErrInvalidEmail):
		return http.StatusBadRequest, "INVALID_EMAIL", "invalid email address"
	case errors.Is(err, ErrVerificationNotFound), errors.Is(err, ErrVerificationExpired):
		return http.StatusBadRequest, "INVALID_TOKEN", "verification token is invalid or expired"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred"
	}
}
