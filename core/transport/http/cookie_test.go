package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/configuration"
)

type CookieSuite struct {
	suite.Suite
}

func (s *CookieSuite) TestNewSecureCookie() {
	tests := []struct {
		name         string
		env          configuration.Environment
		wantSecure   bool
		wantSameSite http.SameSite
	}{
		{
			name:         "development",
			env:          configuration.Development,
			wantSecure:   false,
			wantSameSite: http.SameSiteLaxMode,
		},
		{
			name:         "production",
			env:          configuration.Production,
			wantSecure:   true,
			wantSameSite: http.SameSiteNoneMode,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cookie := NewSecureCookie("session", "abc123", tt.env)

			s.Assert().Equal("session", cookie.Name)
			s.Assert().Equal("abc123", cookie.Value)
			s.Assert().Equal("/", cookie.Path)
			s.Assert().True(cookie.HttpOnly)
			s.Assert().Equal(tt.wantSecure, cookie.Secure)
			s.Assert().Equal(tt.wantSameSite, cookie.SameSite)
		})
	}
}

func TestCookieSuite(t *testing.T) {
	suite.Run(t, new(CookieSuite))
}
