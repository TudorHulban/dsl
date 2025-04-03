package main

import (
	"fmt"
	"strconv"
)

// Parser holds the state of the parsing process.
type Parser struct {
	lex *dslLexer

	tokenCurrent token
	tokenNext    token

	errors []string

	debug bool
}

func NewParser(l *dslLexer) *Parser {
	p := Parser{lex: l}

	p.tokenNext = p.lex.nextToken()
	p.advanceToken()

	if p.debug {
		fmt.Printf("PARSER INIT: Current=%v, Next=%v\n",
			p.tokenCurrent.valueLiteral,
			p.tokenNext.valueLiteral)
	}
	return &p
}

func (p *Parser) advanceToken() {
	p.tokenCurrent = p.tokenNext
	p.tokenNext = p.lex.nextToken()
}

func (p *Parser) EnableDebug() {
	p.debug = true
}

func (p *Parser) logTokenState() {
	if p.debug {
		fmt.Printf("[DEBUG] Current Token: %+v\n", p.tokenCurrent)
		fmt.Printf("[DEBUG] Next Token:    %+v\n", p.tokenNext)
	}
}

func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(
		p.errors,

		fmt.Sprintf(
			"parse error at %s: %s",
			p.tokenCurrent.pos,
			fmt.Sprintf(format, args...),
		),
	)
}

func (p *Parser) tryRecoverAtBlockEnd() {
	if !p.currentTokenIs(tokenRightBrace) && !p.currentTokenIs(tokenEOF) {
		p.advanceToken()
	}
}

type paramsExpect struct {
	Caller       string
	KindExpected tokenKind
}

func (p *Parser) expect(params *paramsExpect) bool {
	if p.tokenCurrent.kind == params.KindExpected {
		p.advanceToken()

		return true
	}

	p.errorf(
		"Caller:%s\nExpected token %v, got %v (%s)",
		params.Caller,
		params.KindExpected,
		p.tokenCurrent.kind,
		p.tokenCurrent.valueLiteral,
	)

	return false
}

// expectIdentifier checks if the current token is an identifier with specific text.
func (p *Parser) expectIdentifier(ident string) bool {
	if p.tokenCurrent.kind == tokenIdentifier && p.tokenCurrent.valueLiteral == ident {
		p.advanceToken()

		return true
	}

	p.errorf(
		"expected identifier '%s', got %v (%s)",
		ident,
		p.tokenCurrent.kind,
		p.tokenCurrent.valueLiteral,
	)

	return false
}

func (p *Parser) parserEntrypoint() *AlertConfiguration {
	var result AlertConfiguration

	// Keep processing until EOF or error
	for {
		if p.tokenCurrent.kind == tokenEOF || p.tokenCurrent.kind == tokenError {
			break
		}

		if p.tokenCurrent.kind == tokenCriteria {
			criteria := p.parseCriteria()
			if criteria != nil {
				result.Criterias = append(result.Criterias, criteria)

				continue // Successfully parsed, move to next token
			}

			// If we get here, parsing failed - skip to next potential dataset
			p.skipToIdentifier("dataset")
			continue
		}

		// Unexpected token - attempt recovery
		if p.tokenCurrent.kind != tokenEOF {
			p.errorf(
				"unexpected token at program root: %v (%s)",
				p.tokenCurrent.kind,
				p.tokenCurrent.valueLiteral,
			)

			p.skipToIdentifier(_dslCriteria)
			continue
		}

		// Safety net - advance if we're stuck
		p.advanceToken()
	}

	return &result
}

func (p *Parser) parseSetting() *Setting {
	var result Setting

	result.Kind = p.tokenCurrent.valueLiteral // "baseline" or "increment"
	p.advanceToken()

	if p.tokenCurrent.kind != tokenIdentifier {
		p.errorf(
			"expected setting name identifier, got %v",
			p.tokenCurrent.kind,
		)

		return nil
	}
	result.Name = p.tokenCurrent.valueLiteral
	p.advanceToken()

	if !p.expect(
		&paramsExpect{
			Caller:       "parseSetting - 1",
			KindExpected: tokenAssign,
		},
	) {
		return nil
	}

	result.Value = p.parseExpression(0) // parse the value expression
	if result.Value == nil {
		p.errorf("invalid setting value expression")
		return nil
	}

	if !p.expect(
		&paramsExpect{
			Caller:       "parseSetting - 2",
			KindExpected: tokenSemicolon,
		},
	) {
		return nil
	}

	return &result
}

// parseExpression - simplified placeholder for expression parsing
// a real implementation needs operator precedence (e.g., Pratt parsing or shunting-yard)
func (p *Parser) parseExpression(precedence int) Expression {
	// very basic: handles literal or variable, optionally followed by operator and another term
	// does not handle precedence or parentheses correctly!
	var left Expression

	switch p.tokenCurrent.kind {
	case tokenNumber:
		// try parsing as float first
		fval, errf := strconv.ParseFloat(p.tokenCurrent.valueLiteral, 64)
		if errf == nil {
			left = newliteral(fval, p.tokenCurrent.valueLiteral)
		} else {
			// try parsing as int
			ival, erri := strconv.Atoi(p.tokenCurrent.valueLiteral)
			if erri == nil {
				left = newliteral(ival, p.tokenCurrent.valueLiteral)
			} else {
				p.errorf("invalid number literal: %s", p.tokenCurrent.valueLiteral)
				return nil
			}
		}

		p.advanceToken()

	case tokenIdentifier:
		left = newvariable(p.tokenCurrent.valueLiteral)

		p.advanceToken()

	default:
		p.errorf("unexpected token in expression: %v (%s)", p.tokenCurrent.kind, p.tokenCurrent.valueLiteral)

		return nil
	}

	// look ahead for a binary operator (super simplified)
	if p.tokenCurrent.kind == tokenOperator {
		op := p.tokenCurrent.valueLiteral

		p.advanceToken()

		right := p.parseExpression(0) // recursive call (doesn't handle precedence)
		if right == nil {
			p.errorf("missing right hand side for operator %s", op)
			return nil
		}

		return newbinaryexpr(left, op, right)
	}

	return left // return just the literal or variable if no operator follows
}

// --- basic error recovery helpers (very naive) ---

func (p *Parser) skipto(types ...tokenKind) {
	for p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		for _, t := range types {
			if p.tokenCurrent.kind == t {
				return // found one of the target types
			}
		}

		p.advanceToken()
	}
}

func (p *Parser) skipToIdentifier(identifiers ...string) {
	for p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		if p.tokenCurrent.kind == tokenIdentifier {
			for _, identifier := range identifiers {
				if p.tokenCurrent.valueLiteral == identifier {
					return // found one of the target idents
				}
			}
		}

		p.advanceToken()
	}
}

func (p *Parser) skipToIdentifierRightBrace(identifiers ...string) {
	for p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		if p.tokenCurrent.kind == tokenRightBrace { // stop at closing brace
			return
		}

		if p.tokenCurrent.kind == tokenIdentifier {
			for _, id := range identifiers {
				if p.tokenCurrent.valueLiteral == id {
					return // found one of the target idents
				}
			}
		}

		p.advanceToken()
	}
}

func (p *Parser) currentTokenIs(t tokenKind) bool {
	return p.tokenCurrent.kind == t
}
