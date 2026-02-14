package http

import (
	"net/http"

	"github.com/bbsbb/go-edge/core/configuration"
)

func NewSecureCookie(name, value string, env configuration.Environment) http.Cookie {
	sameSite := http.SameSiteNoneMode
	secure := true

	if env.IsDevelopment() {
		sameSite = http.SameSiteLaxMode
		secure = false
	}

	return http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
}
