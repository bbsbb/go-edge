package psqlfx

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/domain"
)

type TranslateErrorSuite struct {
	suite.Suite
}

func (s *TranslateErrorSuite) TestNil() {
	s.Assert().Nil(TranslateError(nil))
}

func (s *TranslateErrorSuite) TestNoRows() {
	err := TranslateError(pgx.ErrNoRows)

	s.Require().Error(err)
	s.Assert().ErrorIs(err, domain.ErrNotFound)

	var domErr *domain.Error
	s.Require().ErrorAs(err, &domErr)
	s.Assert().Equal(domain.CodeNotFound, domErr.Code)
	s.Assert().Equal("not found", domErr.Message)
}

func (s *TranslateErrorSuite) TestUniqueViolation() {
	pgErr := &pgconn.PgError{Code: pgUniqueViolation}
	err := TranslateError(pgErr)

	s.Require().Error(err)
	s.Assert().ErrorIs(err, domain.ErrConflict)

	var domErr *domain.Error
	s.Require().ErrorAs(err, &domErr)
	s.Assert().Equal(domain.CodeConflict, domErr.Code)
	s.Assert().Equal("already exists", domErr.Message)
}

func (s *TranslateErrorSuite) TestOtherPgError() {
	pgErr := &pgconn.PgError{Code: "42P01", Message: "relation does not exist"}
	err := TranslateError(pgErr)

	s.Require().Error(err)
	s.Assert().ErrorIs(err, pgErr)

	var domErr *domain.Error
	s.Assert().False(errors.As(err, &domErr))
}

func (s *TranslateErrorSuite) TestGenericError() {
	orig := errors.New("connection refused")
	err := TranslateError(orig)

	s.Require().Error(err)
	s.Assert().ErrorIs(err, orig)

	var domErr *domain.Error
	s.Assert().False(errors.As(err, &domErr))
}

func TestTranslateErrorSuite(t *testing.T) {
	suite.Run(t, new(TranslateErrorSuite))
}
