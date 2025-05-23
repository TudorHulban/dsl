package dslalert

import (
	"errors"
	"fmt"
	"slices"
)

type parser struct {
	lex *dslLexer

	tokenCurrent token
	tokenNext    token

	errors []string

	debug bool
}

type paramsNewParser struct {
	Lexer       *dslLexer
	IsDebugMode bool
}

func newParser(params *paramsNewParser) *parser {
	p := parser{
		lex:   params.Lexer,
		debug: params.IsDebugMode,
	}

	p.tokenNext = p.lex.nextToken()
	p.advanceToken()

	if p.debug {
		fmt.Printf("PARSER INIT: Current=%v, Next=%v\n",
			p.tokenCurrent.valueLiteral,
			p.tokenNext.valueLiteral)
	}

	return &p
}

func (p *parser) advanceToken() {
	p.tokenCurrent = p.tokenNext
	p.tokenNext = p.lex.nextToken()
}

func (p *parser) logTokenState() {
	if p.debug {
		fmt.Printf("[DEBUG] Current Token: %+v\n", p.tokenCurrent)
		fmt.Printf("[DEBUG] Next Token:    %+v\n", p.tokenNext)
	}
}

func (p *parser) errorf(format string, args ...any) {
	p.errors = append(
		p.errors,

		fmt.Sprintf(
			"parse error at %s: %s",
			p.tokenCurrent.pos,
			fmt.Sprintf(format, args...),
		),
	)
}

func (p *parser) tryRecoverAtBlockEnd() {
	if !p.currentTokenIs(tokenRightBrace) && !p.currentTokenIs(tokenEOF) {
		p.advanceToken()
	}
}

type paramsExpect struct {
	Caller       string
	KindExpected tokenKind
}

func (p *parser) expectWTokenAdvance(params *paramsExpect) bool {
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

func (p *parser) expectNoTokenAdvance(params *paramsExpect) bool {
	if p.tokenCurrent.kind == params.KindExpected {
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

func (p *parser) parserEntrypoint() (*AlertConfiguration, error) {
	var result AlertConfiguration

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

			return nil,
				errors.New(
					"there is an error",
				)
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

	return &result,
		nil
}

func (p *parser) skipToIdentifier(identifiers ...string) {
	for p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		if p.tokenCurrent.kind == tokenIdentifier &&
			slices.Contains(
				identifiers,
				p.tokenCurrent.valueLiteral,
			) {
			return // found one of the target identifiers
		}

		p.advanceToken()
	}
}

func (p *parser) skipToIdentifierRightBrace(identifiers ...string) {
	for p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		if p.tokenCurrent.kind == tokenRightBrace {
			return // Stop at closing brace
		}

		if p.tokenCurrent.kind == tokenIdentifier &&
			slices.Contains(identifiers, p.tokenCurrent.valueLiteral) {
			return // Found target identifier
		}

		p.advanceToken()
	}
}

func (p *parser) currentTokenIs(t tokenKind) bool {
	return p.tokenCurrent.kind == t
}
