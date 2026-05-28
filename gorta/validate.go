package gorta

import (
	"strings"
	"unicode"

	"github.com/etornam45/gorta/internals"
)

func validateEmail(email string) error {
	if len(email) == 0 {
		return internals.ErrInvalidEmail
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return internals.ErrInvalidEmail
	}

	local := email[:at]
	domain := email[at+1:]

	if len(local) == 0 || len(domain) == 0 {
		return internals.ErrInvalidEmail
	}

	if !strings.Contains(domain, ".") {
		return internals.ErrInvalidEmail
	}

	if strings.ContainsAny(email, " \t\n\r") {
		return internals.ErrInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return internals.ErrWeakPassword
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
