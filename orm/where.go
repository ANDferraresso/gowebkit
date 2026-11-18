package orm

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Regular expressions shared by all the query builders, compiled once.
const identPattern = `[a-zA-Z_][a-zA-Z0-9_]*`

var (
	identRe          = regexp.MustCompile(`^` + identPattern + `$`)
	identAliasRe     = regexp.MustCompile(`^` + identPattern + ` AS ` + identPattern + `$`)
	qualIdentRe      = regexp.MustCompile(`^` + identPattern + `\.` + identPattern + `$`)
	qualIdentAliasRe = regexp.MustCompile(`^` + identPattern + `\.` + identPattern + ` AS ` + identPattern + `$`)
	setCondRe        = regexp.MustCompile(`^` + identPattern + ` = \?$`)

	// WHERE conditions. "a" = plain column, "b" = qualified column (table.column).
	condValueARe = regexp.MustCompile(`^` + identPattern + ` (<|=|>|<>|<=|>=|LIKE|ILIKE) \?$`)
	condValueBRe = regexp.MustCompile(`^` + identPattern + `\.` + identPattern + ` (<|=|>|<>|<=|>=|LIKE|ILIKE) \?$`)
	condNullARe  = regexp.MustCompile(`^` + identPattern + ` (IS NULL|IS NOT NULL)$`)
	condNullBRe  = regexp.MustCompile(`^` + identPattern + `\.` + identPattern + ` (IS NULL|IS NOT NULL)$`)
	condColARe   = regexp.MustCompile(`^` + identPattern + ` (<|=|>|<>|<=|>=) ` + identPattern + `$`)
	condColBRe   = regexp.MustCompile(`^` + identPattern + `\.` + identPattern + ` (<|=|>|<>|<=|>=) ` + identPattern + `\.` + identPattern + `$`)

	logicOpRe    = regexp.MustCompile(`^(AND|OR|NOT|OR NOT|AND NOT)$`)
	openParenRe  = regexp.MustCompile(`^\(+$`)
	closeParenRe = regexp.MustCompile(`^\)+$`)
)

var errInvalidWhere = errors.New("invalid WHERE clause")

// quoteQualified turns "t.c" into "`t`.`c`".
func quoteQualified(s string) string {
	t, c, _ := strings.Cut(s, ".")
	return fmt.Sprintf("`%s`.`%s`", t, c)
}

// parseWhere validates the WHERE tokens and returns the SQL fragment and binds.
//
// Tokens alternate between: optional "(" groups, a condition (followed by its
// value when the condition ends with "?"), optional ")" groups and a logical
// operator. Example:
//
//	"(", "id = ?", 1, "OR", "name IS NULL", ")", "AND", "c.age > ?", 18
//
// qualified enables conditions on table.column (used by SELECT).
//
// Operators: <, =, >, <>, <=, >=, LIKE, ILIKE. ILIKE is a case-insensitive
// LIKE: it is emitted as is for Postgres and rendered as LIKE for MySQL,
// where LIKE is already case-insensitive with the default (_ci) collations.
func parseWhere(inputs []interface{}, qualified bool) (string, []interface{}, error) {
	var sb strings.Builder
	binds := []interface{}{}
	depth := 0

	const (
		expectCond  = 0 // "(" or a condition
		expectValue = 1 // value for the "?" of the previous condition
		expectOp    = 2 // ")" or a logical operator
	)
	state := expectCond

	for _, in := range inputs {
		switch state {
		case expectCond:
			tok, ok := in.(string)
			if !ok {
				return "", nil, errInvalidWhere
			}
			switch {
			case openParenRe.MatchString(tok):
				sb.WriteString(tok)
				depth += len(tok)
			case condValueARe.MatchString(tok):
				c := strings.Split(tok, " ")
				fmt.Fprintf(&sb, "`%s` %s", c[0], c[1])
				state = expectValue
			case qualified && condValueBRe.MatchString(tok):
				c := strings.Split(tok, " ")
				fmt.Fprintf(&sb, "%s %s", quoteQualified(c[0]), c[1])
				state = expectValue
			case condNullARe.MatchString(tok):
				col, rest, _ := strings.Cut(tok, " ")
				fmt.Fprintf(&sb, "`%s` %s", col, rest)
				state = expectOp
			case qualified && condNullBRe.MatchString(tok):
				col, rest, _ := strings.Cut(tok, " ")
				fmt.Fprintf(&sb, "%s %s", quoteQualified(col), rest)
				state = expectOp
			case condColARe.MatchString(tok):
				c := strings.Split(tok, " ")
				fmt.Fprintf(&sb, "`%s` %s `%s`", c[0], c[1], c[2])
				state = expectOp
			case qualified && condColBRe.MatchString(tok):
				c := strings.Split(tok, " ")
				fmt.Fprintf(&sb, "%s %s %s", quoteQualified(c[0]), c[1], quoteQualified(c[2]))
				state = expectOp
			default:
				return "", nil, errInvalidWhere
			}
		case expectValue:
			sb.WriteString(" ?")
			binds = append(binds, in)
			state = expectOp
		case expectOp:
			tok, ok := in.(string)
			if !ok {
				return "", nil, errInvalidWhere
			}
			switch {
			case closeParenRe.MatchString(tok):
				depth -= len(tok)
				if depth < 0 {
					return "", nil, errInvalidWhere
				}
				sb.WriteString(tok)
			case logicOpRe.MatchString(tok):
				fmt.Fprintf(&sb, " %s ", tok)
				state = expectCond
			default:
				return "", nil, errInvalidWhere
			}
		}
	}

	// The clause must end right after a complete condition (or a ")"),
	// with balanced parentheses.
	if len(inputs) > 0 && (state != expectOp || depth != 0) {
		return "", nil, errInvalidWhere
	}
	return sb.String(), binds, nil
}
