package main

import (
	"fmt"
	"io"
	"strconv"
	"text/scanner"
)

type lexer struct {
	s   scanner.Scanner
	err error
}

func newlexer(r io.Reader) *lexer {
	var s scanner.Scanner
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanStrings | scanner.ScanChars | scanner.ScanComments
	// customize scanner if needed, e.g., operators
	s.IsIdentRune = func(ch rune, i int) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '_') || (i > 0 && ch >= '0' && ch <= '9')
	}
	return &lexer{s: s}
}

func (l *lexer) nexttoken() token {
	if l.err != nil {
		return token{typ: tokenerror, lit: l.err.Error()}
	}
	tok := l.s.Scan()
	pos := l.s.Position
	lit := l.s.TokenText()

	switch tok {
	case scanner.EOF:
		return token{typ: tokeneof, pos: pos}

	case scanner.Ident:
		// could check for keywords here ("dataset", "criteria", etc.)
		return token{typ: tokenident, lit: lit, pos: pos}

	case scanner.String:
		// remove quotes
		unquoted, err := strconv.Unquote(lit)
		if err != nil {
			l.err = fmt.Errorf("invalid string literal at %s: %w", pos, err)
			return token{typ: tokenerror, lit: err.Error(), pos: pos}
		}
		return token{typ: tokenstring, lit: unquoted, pos: pos}

	case scanner.Float, scanner.Int:
		return token{typ: tokennumber, lit: lit, pos: pos}

	case '{':
		return token{typ: tokenlbrace, lit: lit, pos: pos}

	case '}':
		return token{typ: tokenrbrace, lit: lit, pos: pos}

	case '=':
		return token{typ: tokenassign, lit: lit, pos: pos}

	case ';':
		return token{typ: tokensemicolon, lit: lit, pos: pos}

		// very basic operator handling - needs improvement for multi-char ops (>=)
	case '>', '<', '+', '-', '*', '/':
		// peek ahead for multi-char operators like >=, <=, ==, !=
		next := l.s.Peek()
		if lit == ">" && next == '=' {
			l.s.Scan() // consume '='
			return token{typ: tokenoperator, lit: ">=", pos: pos}
		}
		if lit == "<" && next == '=' {
			l.s.Scan() // consume '='
			return token{typ: tokenoperator, lit: "<=", pos: pos}
		}
		if lit == "=" && next == '=' {
			l.s.Scan() // consume '='
			return token{typ: tokenoperator, lit: "==", pos: pos}
		}
		if lit == "!" && next == '=' {
			l.s.Scan() // consume '='
			return token{typ: tokenoperator, lit: "!=", pos: pos}
		}
		return token{typ: tokenoperator, lit: lit, pos: pos}

	default:
		// handle other chars or report error
		l.err = fmt.Errorf("unexpected character '%s' at %s", lit, pos)

		return token{typ: tokenerror, lit: l.err.Error(), pos: pos}
	}
}
