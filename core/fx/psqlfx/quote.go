package psqlfx

import (
	"strings"

	"github.com/jackc/pgx/v5"
)

func QuoteIdentifier(ident []string) string {
	var pgxIdent pgx.Identifier = ident
	return pgxIdent.Sanitize()
}

func QuoteLiteral(value string) string {
	value = strings.ReplaceAll(value, `'`, `''`)

	if strings.Contains(value, `\`) {
		value = strings.ReplaceAll(value, `\`, `\\`)
		return ` E'` + value + `'`
	}
	return `'` + value + `'`
}
