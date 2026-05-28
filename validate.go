package gorta

import (
	"strings"
	"unicode"
)

func validateEmail(email string) error {
	if len(email) == 0 {
		return ErrInvalidEmail
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return ErrInvalidEmail
	}

	local := email[:at]
	domain := email[at+1:]

	if len(local) == 0 || len(domain) == 0 {
		return ErrInvalidEmail
	}

	if !strings.Contains(domain, ".") {
		return ErrInvalidEmail
	}

	if strings.ContainsAny(email, " \t\n\r") {
		return ErrInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
}