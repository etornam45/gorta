package core

import (
	"net/http"
	"strings"
)

const DefaultSessionCookie = "gorta_session"

func ExtractToken(r *http.Request, cookieName string) string {
	if cookieName == "" {
		cookieName = DefaultSessionCookie
	}
	if cookie, err := r.Cookie(cookieName); err == nil {
		return cookie.Value
	}
	if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	return ""
}

func SessionMetaFromRequest(r *http.Request) SessionMeta {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return SessionMeta{
		IPAddress: ip,
		UserAgent: r.Header.Get("User-Agent"),
	}
}