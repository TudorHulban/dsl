package main

import (
	"fmt"
	"io"
	"strconv"
	"text/scanner"
)

type dslLexer struct {
	s            scanner.Scanner
	errorParsing error
}

func newlexer(r io.Reader) *dslLexer {
	var s scanner.Scanner
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanStrings | scanner.ScanChars | scanner.ScanComments

	// customize scanner if needed, e.g., operators
	s.IsIdentRune = func(ch rune, i int) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '_') || (i > 0 && ch >= '0' && ch <= '9')
	}

	return &dslLexer{s: s}
}

func (l *dslLexer) nextToken() token {
	if l.errorParsing != nil {
		return token{kind: tokenError, valueLiteral: l.errorParsing.Error()}
	}
	tok := l.s.Scan()
	pos := l.s.Position
	lit := l.s.TokenText()

	switch tok {
	case scanner.EOF:
		return token{kind: tokenEOF, pos: pos}

	case scanner.Ident:
		// could check for keywords here ("dataset", "criteria", etc.)
		return token{kind: tokenIdentifier, valueLiteral: lit, pos: pos}

	case scanner.String:
		// remove quotes
		unquoted, err := strconv.Unquote(lit)
		if err != nil {
			l.errorParsing = fmt.Errorf("invalid string literal at %s: %w", pos, err)
			return token{kind: tokenError, valueLiteral: err.Error(), pos: pos}
		}
		return token{kind: tokenStringLiteral, valueLiteral: unquoted, pos: pos}

	case scanner.Float, scanner.Int:
		return token{kind: tokenNumber, valueLiteral: lit, pos: pos}

	case '{':
		return token{kind: tokenLeftBrace, valueLiteral: lit, pos: pos}

	case '}':
		return token{kind: tokenRightBrace, valueLiteral: lit, pos: pos}

	case '=':
		return token{kind: tokenAssign, valueLiteral: lit, pos: pos}

	case ';':
		return token{kind: tokenSemicolon, valueLiteral: lit, pos: pos}

		// very basic operator handling - needs improvement for multi-char ops (>=)
	case '>', '<', '+', '-', '*', '/':
		// peek ahead for multi-char operators like >=, <=, ==, !=
		next := l.s.Peek()
		if lit == ">" && next == '=' {
			l.s.Scan() // consume '='
			return token{kind: tokenOperator, valueLiteral: ">=", pos: pos}
		}
		if lit == "<" && next == '=' {
			l.s.Scan() // consume '='
			return token{kind: tokenOperator, valueLiteral: "<=", pos: pos}
		}
		if lit == "=" && next == '=' {
			l.s.Scan() // consume '='
			return token{kind: tokenOperator, valueLiteral: "==", pos: pos}
		}
		if lit == "!" && next == '=' {
			l.s.Scan() // consume '='
			return token{kind: tokenOperator, valueLiteral: "!=", pos: pos}
		}
		return token{kind: tokenOperator, valueLiteral: lit, pos: pos}

	default:
		// handle other chars or report error
		l.errorParsing = fmt.Errorf("unexpected character '%s' at %s", lit, pos)

		return token{kind: tokenError, valueLiteral: l.errorParsing.Error(), pos: pos}
	}
}
