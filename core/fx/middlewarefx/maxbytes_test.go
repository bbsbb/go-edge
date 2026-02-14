package middlewarefx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MaxBytesSuite struct {
	suite.Suite
}

func (s *MaxBytesSuite) TestRejectsOversizedBody() {
	handler := MaxBytes(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 100)))
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusRequestEntityTooLarge, rec.Code)
}

func (s *MaxBytesSuite) TestAllowsBodyWithinLimit() {
	handler := MaxBytes(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
}

func (s *MaxBytesSuite) TestDefaultsToOneMBWhenZero() {
	handler := MaxBytes(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small body"))
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
}

func TestMaxBytesSuite(t *testing.T) {
	suite.Run(t, new(MaxBytesSuite))
}
