package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/domain"
	coretesting "github.com/bbsbb/go-edge/core/testing"
)

var noopLogger = coretesting.NewNoopLogger()

type WriteErrorSuite struct {
	suite.Suite
}

// requestWithReqID runs the request through chi's RequestID middleware
// so that chimw.GetReqID returns a value from the context.
func (s *WriteErrorSuite) requestWithReqID(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)

	var captured *http.Request
	handler := chimw.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)
	return rr, captured
}

func (s *WriteErrorSuite) TestWriteError() {
	inner := errors.Join(fmt.Errorf("field A is required"), fmt.Errorf("field B must be positive"))

	tests := []struct {
		name       string
		err        error
		withReqID  bool
		method     string
		path       string
		wantStatus int
		wantCode   string
		wantDetail string
		wantErrors []string
		wantReqID  bool
	}{
		{
			name:       "domain error with known code",
			err:        domain.NewError(domain.CodeNotFound, "user not found"),
			withReqID:  true,
			method:     http.MethodGet,
			path:       "/users/123",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
			wantDetail: "user not found",
			wantReqID:  true,
		},
		{
			name:       "domain error with unknown code",
			err:        domain.NewError(domain.Code("CUSTOM"), "custom error"),
			withReqID:  false,
			method:     http.MethodGet,
			path:       "/things/456",
			wantStatus: http.StatusInternalServerError,
			wantCode:   "CUSTOM",
			wantDetail: "custom error",
			wantReqID:  false,
		},
		{
			name:       "domain error with multi-error details",
			err:        domain.WrapError(domain.CodeValidation, "validation failed", inner),
			withReqID:  true,
			method:     http.MethodPost,
			path:       "/users",
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION",
			wantDetail: "validation failed",
			wantErrors: []string{"field A is required", "field B must be positive"},
			wantReqID:  true,
		},
		{
			name:       "non-domain error",
			err:        fmt.Errorf("something went wrong"),
			withReqID:  false,
			method:     http.MethodGet,
			path:       "/secret/path",
			wantStatus: http.StatusInternalServerError,
			wantCode:   "",
			wantDetail: "",
			wantReqID:  false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var rr *httptest.ResponseRecorder
			var req *http.Request
			if tt.withReqID {
				rr, req = s.requestWithReqID(tt.method, tt.path)
			} else {
				rr = httptest.NewRecorder()
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			WriteError(rr, req, tt.err, noopLogger)

			s.Assert().Equal(tt.wantStatus, rr.Code)
			s.Assert().Equal("application/problem+json", rr.Header().Get("Content-Type"))

			var resp ErrorShape
			s.Require().NoError(json.NewDecoder(rr.Body).Decode(&resp))
			s.Assert().Equal(tt.wantStatus, resp.Status)
			s.Assert().Equal(tt.wantCode, resp.Code)
			s.Assert().Equal(tt.wantDetail, resp.Detail)
			s.Assert().Equal(tt.path, resp.Instance)

			if tt.wantReqID {
				s.Assert().NotEmpty(resp.RequestID)
			} else {
				s.Assert().Empty(resp.RequestID)
			}

			if len(tt.wantErrors) > 0 {
				s.Assert().Len(resp.Errors, len(tt.wantErrors))
				for _, e := range tt.wantErrors {
					s.Assert().Contains(resp.Errors, e)
				}
			} else {
				s.Assert().Empty(resp.Errors)
			}
		})
	}
}

func (s *WriteErrorSuite) TestAllKnownCodeMappings() {
	tests := []struct {
		code   domain.Code
		status int
	}{
		{domain.CodeNotFound, http.StatusNotFound},
		{domain.CodeConflict, http.StatusConflict},
		{domain.CodeValidation, http.StatusBadRequest},
		{domain.CodeForbidden, http.StatusForbidden},
		{domain.CodeInvariant, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		s.Run(string(tt.code), func() {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			WriteError(rr, req, domain.NewError(tt.code, "test"), noopLogger)

			s.Assert().Equal(tt.status, rr.Code)

			var resp ErrorShape
			s.Require().NoError(json.NewDecoder(rr.Body).Decode(&resp))
			s.Assert().Equal(tt.status, resp.Status)
		})
	}
}

func TestWriteErrorSuite(t *testing.T) {
	suite.Run(t, new(WriteErrorSuite))
}
