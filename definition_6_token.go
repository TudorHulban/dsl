package main

import "text/scanner"

type tokenKind int

const (
	tokenEOF tokenKind = iota + 1
	tokenError
	tokenIdentifier // identifier like 'criteria', 'value', 'my_col'
	tokenCriteria
	tokenMonitor
	tokenLevel
	tokenWhen
	tokenStringLiteral // "string literal"
	tokenNumber        // 123, 50.5
	tokenLeftBrace     // {
	tokenRightBrace    // }
	tokenAssign        // =
	tokenSemicolon     // ;
	tokenOperator      // >, >=, <, <=, ==, !=, +, -, *, /
)

const (
	_dslCriteria = "criteria"
	_dslMonitor  = "monitor"
	_dslLevel    = "level"
	_dslWhen     = "when"
)

// token represents a single token from the input.
type token struct {
	kind         tokenKind
	valueLiteral string // literal value of the token (e.g., "my_col", ">=", "100")
	pos          scanner.Position
}
