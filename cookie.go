package gorta

import (
	"net/http"
	"time"
)

func (a *Auth) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.config.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   a.config.CookieDomain,
		Expires:  time.Now().Add(a.config.SessionDuration),
		MaxAge:   int(a.config.SessionDuration.Seconds()),
		Secure:   a.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *Auth) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.config.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   a.config.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   a.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
