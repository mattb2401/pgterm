package pgterm

import (
	"strings"
	"unicode"
)

// Common SQL keywords to uppercase
var sqlKeywords = map[string]bool{
	"select": true, "from": true, "join": true, "on": true, "where": true,
	"update": true, "set": true, "insert": true, "into": true, "values": true,
	"delete": true, "create": true, "table": true, "as": true, "and": true,
	"or": true, "not": true, "null": true, "inner": true, "left": true, "right": true,
	"outer": true, "group": true, "by": true, "order": true, "limit": true,
}

// addSchema prefixes table names with schema if not already qualified
func addSchema(sql, schema string) string {
	tokens := strings.Fields(sql)
	keywords := map[string]bool{
		"FROM": true, "JOIN": true,
		"UPDATE": true, "INTO": true,
		"DELETE": true, "TABLE": true,
	}

	for i := 0; i < len(tokens); i++ {
		upper := strings.ToUpper(tokens[i])
		if keywords[upper] {
			// get next token as table name if exists
			if i+1 < len(tokens) {
				tbl := tokens[i+1]

				// strip alias punctuation
				tblClean := strings.TrimFunc(tbl, func(r rune) bool {
					return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.'
				})

				// skip if already schema-qualified or subquery
				if !strings.Contains(tblClean, ".") && !strings.HasPrefix(tblClean, "(") {
					tokens[i+1] = strings.Replace(tokens[i+1], tblClean, schema+"."+tblClean, 1)
				}
			}
		}
	}
	// Capitalize keywords
	for i := range tokens {
		if sqlKeywords[strings.ToLower(tokens[i])] {
			tokens[i] = strings.ToUpper(tokens[i])
		}
	}

	return strings.Join(tokens, " ")
}

// tokenize splits SQL into words/punctuation
func tokenize(sql string) []string {
	var tokens []string
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}
	}
	for _, r := range sql {
		if unicode.IsSpace(r) {
			flush()
		} else if strings.ContainsRune("(),;=<>", r) {
			flush()
			tokens = append(tokens, string(r))
		} else {
			buf.WriteRune(r)
		}
	}
	flush()
	return tokens
}

// isIdentifier checks if a token is an unquoted identifier
func isIdentifier(tok string) bool {
	if tok == "" {
		return false
	}
	// quoted identifiers are fine too
	if strings.HasPrefix(tok, `"`) && strings.HasSuffix(tok, `"`) {
		return true
	}
	// must start with a letter or underscore
	r := rune(tok[0])
	return unicode.IsLetter(r) || r == '_'
}
