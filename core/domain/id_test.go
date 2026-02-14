package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type IDSuite struct {
	suite.Suite
}

func (s *IDSuite) TestNewID_NonZero() {
	id := NewID()
	s.Assert().False(id.IsZero())
}

func (s *IDSuite) TestParseID_RoundTrips() {
	original := NewID()
	parsed, err := ParseID(original.String())
	s.Require().NoError(err)
	s.Assert().Equal(original, parsed)
}

func (s *IDSuite) TestParseID_GarbageReturnsValidationError() {
	_, err := ParseID("not-a-uuid")
	s.Require().Error(err)
	s.Assert().ErrorIs(err, ErrValidation)
}

func (s *IDSuite) TestIDFrom_WrapsUUID() {
	u := uuid.Must(uuid.NewV7())
	id := IDFrom(u)
	s.Assert().Equal(u, id.UUID())
}

func (s *IDSuite) TestIsZero_OnZeroValue() {
	var id ID
	s.Assert().True(id.IsZero())
}

func TestIDSuite(t *testing.T) {
	suite.Run(t, new(IDSuite))
}
