package psqlfx

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QuoteSuite struct {
	suite.Suite
}

func (s *QuoteSuite) TestQuoteIdentifier() {
	tests := []struct {
		name  string
		ident []string
		want  string
	}{
		{
			name:  "single identifier",
			ident: []string{"users"},
			want:  `"users"`,
		},
		{
			name:  "schema and table",
			ident: []string{"app", "users"},
			want:  `"app"."users"`,
		},
		{
			name:  "identifier with special chars",
			ident: []string{"my-schema", "my-table"},
			want:  `"my-schema"."my-table"`,
		},
		{
			name:  "identifier with quotes",
			ident: []string{`my"table`},
			want:  `"my""table"`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := QuoteIdentifier(tt.ident)
			s.Assert().Equal(tt.want, got)
		})
	}
}

func (s *QuoteSuite) TestQuoteLiteral() {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "simple string",
			value: "hello",
			want:  `'hello'`,
		},
		{
			name:  "string with single quote",
			value: "it's",
			want:  `'it''s'`,
		},
		{
			name:  "string with backslash",
			value: `path\to\file`,
			want:  ` E'path\\to\\file'`,
		},
		{
			name:  "string with both",
			value: `it's a \path`,
			want:  ` E'it''s a \\path'`,
		},
		{
			name:  "empty string",
			value: "",
			want:  `''`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := QuoteLiteral(tt.value)
			s.Assert().Equal(tt.want, got)
		})
	}
}

func TestQuoteSuite(t *testing.T) {
	suite.Run(t, new(QuoteSuite))
}
