package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorSuite struct {
	suite.Suite
}

func (s *ErrorSuite) TestError_ErrorWithoutWrapped() {
	err := NewError(CodeNotFound, "thing not found")
	s.Assert().Equal("NOT_FOUND: thing not found", err.Error())
}

func (s *ErrorSuite) TestError_ErrorWithWrapped() {
	inner := fmt.Errorf("db connection failed")
	err := WrapError(CodeNotFound, "thing not found", inner)
	s.Assert().Equal("NOT_FOUND: thing not found: db connection failed", err.Error())
}

func (s *ErrorSuite) TestError_IsSameCode() {
	err1 := NewError(CodeNotFound, "first")
	err2 := NewError(CodeNotFound, "second")
	s.Assert().True(errors.Is(err1, err2))
}

func (s *ErrorSuite) TestError_IsDifferentCode() {
	err1 := NewError(CodeNotFound, "not found")
	err2 := NewError(CodeConflict, "conflict")
	s.Assert().False(errors.Is(err1, err2))
}

func (s *ErrorSuite) TestError_IsNonDomainError() {
	err := NewError(CodeNotFound, "not found")
	s.Assert().False(errors.Is(err, fmt.Errorf("plain error")))
}

func (s *ErrorSuite) TestError_UnwrapReturnsWrappedError() {
	inner := fmt.Errorf("inner")
	err := WrapError(CodeNotFound, "outer", inner)
	s.Assert().Equal(inner, errors.Unwrap(err))
}

func (s *ErrorSuite) TestError_UnwrapReturnsNilWhenNone() {
	err := NewError(CodeNotFound, "no wrapper")
	s.Assert().Nil(errors.Unwrap(err))
}

func (s *ErrorSuite) TestNewError() {
	err := NewError(CodeValidation, "invalid input")
	s.Assert().Equal(CodeValidation, err.Code)
	s.Assert().Equal("invalid input", err.Message)
	s.Assert().Nil(err.Err)
}

func (s *ErrorSuite) TestWrapError() {
	inner := fmt.Errorf("cause")
	err := WrapError(CodeConflict, "already exists", inner)
	s.Assert().Equal(CodeConflict, err.Code)
	s.Assert().Equal("already exists", err.Message)
	s.Assert().Equal(inner, err.Err)
}

func (s *ErrorSuite) TestSentinelErrors() {
	tests := []struct {
		name     string
		sentinel *Error
		code     Code
	}{
		{"ErrNotFound", ErrNotFound, CodeNotFound},
		{"ErrConflict", ErrConflict, CodeConflict},
		{"ErrValidation", ErrValidation, CodeValidation},
		{"ErrForbidden", ErrForbidden, CodeForbidden},
		{"ErrInvariant", ErrInvariant, CodeInvariant},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := NewError(tt.code, "test message")
			s.Assert().True(errors.Is(err, tt.sentinel))
		})
	}
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorSuite))
}
