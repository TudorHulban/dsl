package main

import "text/scanner"

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenError
	tokenIdentifier    // identifier like 'dataset', 'criteria', 'value', 'my_var'
	tokenStringLiteral // "string literal"
	tokenNumber        // 123, 50.5
	tokenLeftBrace     // {
	tokenRightBrace    // }
	tokenAssign        // =
	tokenSemicolon     // ;
	tokenOperator      // >, >=, <, <=, ==, !=, +, -, *, /
)

// token represents a single token from the input.
type token struct {
	kind         tokenKind
	valueLiteral string // literal value of the token (e.g., "dataset", "my_col", ">=", "100")
	pos          scanner.Position
}
