package main

import "text/scanner"

// tokentype represents the type of token found by the lexer.
type tokentype int

const (
	tokeneof tokentype = iota
	tokenerror
	tokenident     // identifier like 'dataset', 'criteria', 'value', 'my_var'
	tokenstring    // "string literal"
	tokennumber    // 123, 50.5
	tokenlbrace    // {
	tokenrbrace    // }
	tokenassign    // =
	tokensemicolon // ;
	tokenoperator  // >, >=, <, <=, ==, !=, +, -, *, /
	// keywords could be specific tokens or handled by checking tokenident text
)

// token represents a single token from the input.
type token struct {
	typ tokentype
	lit string // literal value of the token (e.g., "dataset", "my_col", ">=", "100")
	pos scanner.Position
}
