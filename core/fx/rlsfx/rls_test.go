package rlsfx

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type RLSSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	db   *DB
}

func (s *RLSSuite) SetupSuite() {
	dsn := "host=localhost port=5432 user=root password=root dbname=test_core sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	s.Require().NoError(err)

	s.Require().NoError(pool.Ping(context.Background()))

	s.pool = pool
	s.db = &DB{
		pool:   pool,
		schema: "app",
		field:  "current_organization",
		logger: coretesting.NewNoopLogger(),
	}
}

func (s *RLSSuite) TearDownSuite() {
	s.pool.Close()
}

func (s *RLSSuite) ctxWithOrg(orgID uuid.UUID) context.Context {
	return domain.ContextWithOrganization(context.Background(), &domain.Organization{
		ID:   orgID,
		Slug: "test-org",
	})
}

func (s *RLSSuite) TestTx_SetsRLSVariable() {
	orgID := uuid.Must(uuid.NewV7())
	ctx := s.ctxWithOrg(orgID)

	err := s.db.Tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var value string
		err := tx.QueryRow(ctx, "SELECT current_setting('app.current_organization')").Scan(&value)
		s.Require().NoError(err)
		s.Assert().Equal(orgID.String(), value)
		return nil
	})
	s.Require().NoError(err)
}

func (s *RLSSuite) TestTx_VariableClearedAfterCommit() {
	orgID := uuid.Must(uuid.NewV7())
	ctx := s.ctxWithOrg(orgID)

	err := s.db.Tx(ctx, func(_ context.Context, _ pgx.Tx) error {
		return nil
	})
	s.Require().NoError(err)

	// After commit, a new connection should not see the variable.
	conn, err := s.pool.Acquire(context.Background())
	s.Require().NoError(err)
	defer conn.Release()

	var value string
	err = conn.QueryRow(context.Background(), "SELECT current_setting('app.current_organization', true)").Scan(&value)
	s.Require().NoError(err)
	s.Assert().Empty(value)
}

func (s *RLSSuite) TestTx_MissingOrganization() {
	err := s.db.Tx(context.Background(), func(_ context.Context, _ pgx.Tx) error {
		s.Fail("fn should not be called")
		return nil
	})
	s.Assert().ErrorIs(err, domain.ErrMissingOrganization)
}

func (s *RLSSuite) TestTx_NestedTransaction() {
	orgA := uuid.Must(uuid.NewV7())
	orgB := uuid.Must(uuid.NewV7())
	ctxA := s.ctxWithOrg(orgA)

	err := s.db.Tx(ctxA, func(ctx context.Context, tx pgx.Tx) error {
		// Verify outer tx sees org A.
		var value string
		err := tx.QueryRow(ctx, "SELECT current_setting('app.current_organization')").Scan(&value)
		s.Require().NoError(err)
		s.Assert().Equal(orgA.String(), value)

		// Inner Tx with org B — should create a savepoint.
		ctxB := domain.ContextWithOrganization(ctx, &domain.Organization{ID: orgB, Slug: "org-b"})
		err = s.db.Tx(ctxB, func(innerCtx context.Context, innerTx pgx.Tx) error {
			var innerValue string
			err := innerTx.QueryRow(innerCtx, "SELECT current_setting('app.current_organization')").Scan(&innerValue)
			s.Require().NoError(err)
			s.Assert().Equal(orgB.String(), innerValue)
			return nil
		})
		s.Require().NoError(err)

		// After inner tx commits, outer tx should see org A again
		// (SET LOCAL respects savepoint boundaries).
		err = tx.QueryRow(ctx, "SELECT current_setting('app.current_organization')").Scan(&value)
		s.Require().NoError(err)
		s.Assert().Equal(orgA.String(), value)

		return nil
	})
	s.Require().NoError(err)
}

func (s *RLSSuite) TestTx_NestedRollback() {
	orgA := uuid.Must(uuid.NewV7())
	orgB := uuid.Must(uuid.NewV7())
	ctxA := s.ctxWithOrg(orgA)
	errInner := errors.New("inner failure")

	err := s.db.Tx(ctxA, func(ctx context.Context, tx pgx.Tx) error {
		// Inner Tx with org B — will fail.
		ctxB := domain.ContextWithOrganization(ctx, &domain.Organization{ID: orgB, Slug: "org-b"})
		err := s.db.Tx(ctxB, func(_ context.Context, _ pgx.Tx) error {
			return errInner
		})
		s.Assert().ErrorIs(err, errInner)

		// Outer tx should still see org A after inner rollback.
		var value string
		err = tx.QueryRow(ctx, "SELECT current_setting('app.current_organization')").Scan(&value)
		s.Require().NoError(err)
		s.Assert().Equal(orgA.String(), value)

		return nil
	})
	s.Require().NoError(err)
}

func (s *RLSSuite) TestTx_ErrorPropagation() {
	orgID := uuid.Must(uuid.NewV7())
	ctx := s.ctxWithOrg(orgID)
	expectedErr := errors.New("something went wrong")

	err := s.db.Tx(ctx, func(_ context.Context, _ pgx.Tx) error {
		return expectedErr
	})
	s.Assert().ErrorIs(err, expectedErr)
}

func (s *RLSSuite) TestTx_ContextCarriesTx() {
	orgID := uuid.Must(uuid.NewV7())
	ctx := s.ctxWithOrg(orgID)

	err := s.db.Tx(ctx, func(ctx context.Context, _ pgx.Tx) error {
		// The tx should be stored in context by Tx().
		tx := psqlfx.TxFromContext(ctx)
		s.Require().NotNil(tx)

		// The tx from context should be usable.
		var value string
		err := tx.QueryRow(ctx, "SELECT current_setting('app.current_organization')").Scan(&value)
		s.Require().NoError(err)
		s.Assert().Equal(orgID.String(), value)

		return nil
	})
	s.Require().NoError(err)
}

func TestRLSSuite(t *testing.T) {
	suite.Run(t, new(RLSSuite))
}
