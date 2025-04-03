package main

import (
	"fmt"
	"io"
	"strconv"
	"text/scanner"
)

type dslLexer struct {
	scaner       scanner.Scanner
	errorParsing error
}

func newLexer(reader io.Reader) *dslLexer {
	var s scanner.Scanner

	s.Init(reader)
	s.Mode = scanner.ScanIdents |
		scanner.ScanFloats |
		scanner.ScanStrings |
		scanner.ScanChars |
		scanner.ScanComments

	// customize scanner if needed, e.g., operators
	s.IsIdentRune = func(ch rune, i int) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '_') || (i > 0 && ch >= '0' && ch <= '9')
	}

	return &dslLexer{
		scaner: s,
	}
}

func (l *dslLexer) nextToken() token {
	if l.errorParsing != nil {
		return token{
			kind:         tokenError,
			valueLiteral: l.errorParsing.Error(),
		}
	}

	currentToken := l.scaner.Scan()
	literalToken := l.scaner.TokenText()
	position := l.scaner.Position

	switch currentToken {
	case scanner.EOF:
		return token{
			kind: tokenEOF,
			pos:  position,
		}

	case scanner.Ident:
		switch literalToken {
		case _dslCriteria:
			return token{
				kind:         tokenCriteria,
				valueLiteral: literalToken,
				pos:          position,
			}

		case _dslMonitor:
			return token{
				kind:         tokenMonitor,
				valueLiteral: literalToken,
				pos:          position,
			}

		case _dslLevel:
			return token{
				kind:         tokenLevel,
				valueLiteral: literalToken,
				pos:          position,
			}

		case _dslWhen:
			return token{
				kind:         tokenWhen,
				valueLiteral: literalToken,
				pos:          position,
			}

		default:
			return token{
				kind:         tokenIdentifier,
				valueLiteral: literalToken,
				pos:          position,
			}
		}

	case scanner.String:
		unquoted, err := strconv.Unquote(literalToken)
		if err != nil {
			l.errorParsing = fmt.Errorf(
				"invalid string literal at %s: %w",
				position,
				err,
			)

			return token{
				kind:         tokenError,
				valueLiteral: err.Error(),
				pos:          position,
			}
		}

		return token{
			kind:         tokenStringLiteral,
			valueLiteral: unquoted,
			pos:          position,
		}

	case scanner.Float, scanner.Int:
		return token{
			kind:         tokenNumber,
			valueLiteral: literalToken,
			pos:          position,
		}

	case '{':
		return token{
			kind:         tokenLeftBrace,
			valueLiteral: literalToken,
			pos:          position,
		}

	case '}':
		return token{
			kind:         tokenRightBrace,
			valueLiteral: literalToken,
			pos:          position,
		}

	case '=':
		return token{
			kind:         tokenAssign,
			valueLiteral: literalToken,
			pos:          position,
		}

	case ';':
		return token{
			kind:         tokenSemicolon,
			valueLiteral: literalToken,
			pos:          position,
		}

		// very basic operator handling - needs improvement for multi-char ops (>=)
	case '>', '<', '+', '-', '*', '/':
		// peek ahead for multi-char operators like >=, <=, ==, !=
		next := l.scaner.Peek()
		if literalToken == ">" && next == '=' {
			l.scaner.Scan() // consume '='

			return token{
				kind:         tokenOperator,
				valueLiteral: ">=",
				pos:          position,
			}
		}

		if literalToken == "<" && next == '=' {
			l.scaner.Scan() // consume '='

			return token{
				kind:         tokenOperator,
				valueLiteral: "<=",
				pos:          position,
			}
		}

		if literalToken == "=" && next == '=' {
			l.scaner.Scan() // consume '='

			return token{
				kind:         tokenOperator,
				valueLiteral: "==",
				pos:          position,
			}
		}

		if literalToken == "!" && next == '=' {
			l.scaner.Scan() // consume '='

			return token{
				kind:         tokenOperator,
				valueLiteral: "!=",
				pos:          position,
			}
		}

		return token{
			kind:         tokenOperator,
			valueLiteral: literalToken,
			pos:          position,
		}

	default:
		// handle other chars or report error
		l.errorParsing = fmt.Errorf(
			"unexpected character '%s' at %s",
			literalToken,
			position,
		)

		return token{
			kind:         tokenError,
			valueLiteral: l.errorParsing.Error(),
			pos:          position,
		}
	}
}
