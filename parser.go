package main

import (
	"bytes"
	"fmt"
	"io"
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

// errorf records a parsing error.
func (p *Parser) errorf(format string, args ...any) {
	msg := fmt.Sprintf("parse error at %s: %s", p.tokenCurrent.pos, fmt.Sprintf(format, args...))

	p.errors = append(p.errors, msg)
}

// expect checks if the current token matches the expected type.
func (p *Parser) expect(t tokenKind) bool {
	if p.tokenCurrent.kind == t {
		p.advanceToken()

		return true
	}

	p.errorf(
		"expected token %v, got %v (%s)",
		t,
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

// --- parsing functions (recursive descent) ---

// func (p *Parser) parseDataset() *dataset {
// 	// 1. Verify 'dataset' keyword
// 	if !p.expectIdentifier("dataset") {
// 		p.errorf("expected 'dataset' keyword")

// 		return nil
// 	}

// 	// 2. Get dataset name (string literal)
// 	if p.tokenCurrent.kind != tokenStringLiteral {
// 		p.errorf(
// 			"expected dataset name string, got %v (%s)",
// 			p.tokenCurrent.kind, p.tokenCurrent.valueLiteral,
// 		)

// 		return nil
// 	}

// 	ds := dataset{
// 		Name: strings.Trim(p.tokenCurrent.valueLiteral, `"`), // Remove quotes
// 	}
// 	p.advanceToken()

// 	// 3. Verify opening brace
// 	if !p.expect(tokenLeftBrace) {
// 		p.errorf("expected '{' after dataset name")

// 		return nil
// 	}

// 	// 4. Parse criteria inside dataset
// 	for p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF {
// 		switch {
// 		case p.tokenCurrent.kind == tokenIdentifier && p.tokenCurrent.valueLiteral == "criteria":
// 			crit := p.parseCriteria()
// 			if crit != nil {
// 				ds.Criteria = append(ds.Criteria, crit)
// 			} else {
// 				p.skiptoidentorbrace("criteria") // Error recovery
// 			}

// 		default:
// 			p.errorf(
// 				"unexpected token in dataset block: %v (%s)",
// 				p.tokenCurrent.kind, p.tokenCurrent.valueLiteral,
// 			)

// 			p.advanceToken()
// 		}
// 	}

// 	// 5. Verify closing brace
// 	if !p.expect(tokenRightBrace) {
// 		p.errorf("dataset block not properly closed")

// 		return nil
// 	}

// 	return &ds
// }

func (p *Parser) parserEntrypoint() *AlertConfiguration {
	var result AlertConfiguration

	// Keep processing until EOF or error
	for {
		// Check for termination conditions first
		if p.tokenCurrent.kind == tokenEOF || p.tokenCurrent.kind == tokenError {
			break
		}

		if p.tokenCurrent.kind == tokenIdentifier && p.tokenCurrent.valueLiteral == _dslCriteria {
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

func (p *Parser) parseCriteria() *Criteria {
	var result Criteria

	if !p.expectIdentifier("criteria") {
		return nil
	}

	if p.tokenCurrent.kind != tokenStringLiteral {
		p.errorf("expected criteria name string, got %v", p.tokenCurrent.kind)
		return nil
	}

	result.Name = p.tokenCurrent.valueLiteral
	p.advanceToken()

	if !p.expect(tokenLeftBrace) {
		return nil
	}

	for p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		switch {
		case p.tokenCurrent.kind == tokenIdentifier && (p.tokenCurrent.valueLiteral == "baseline" || p.tokenCurrent.valueLiteral == "increment"):
			sett := p.parseSetting()
			if sett != nil {
				result.Settings = append(result.Settings, sett)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skipToIdentifierRightBrace("baseline", "increment", "monitor")
			}

		case p.tokenCurrent.kind == tokenIdentifier && p.tokenCurrent.valueLiteral == "monitor":
			mon := p.parseMonitor()
			if mon != nil {
				result.Monitors = append(result.Monitors, mon)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skipToIdentifierRightBrace("baseline", "increment", "monitor")
			}

		default:
			p.errorf("unexpected token inside criteria block: %v (%s)", p.tokenCurrent.kind, p.tokenCurrent.valueLiteral)
			p.skipToIdentifierRightBrace("baseline", "increment", "monitor")
		}
	}

	if !p.expect(tokenRightBrace) {
		p.errorf("criteria block not properly closed")
		if p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF {
			p.advanceToken()
		}
	}

	return &result
}

func (p *Parser) parseSetting() *Setting {
	var result Setting

	result.Kind = p.tokenCurrent.valueLiteral // "baseline" or "increment"
	p.advanceToken()

	if p.tokenCurrent.kind != tokenIdentifier {
		p.errorf("expected setting name identifier, got %v", p.tokenCurrent.kind)
		return nil
	}
	result.Name = p.tokenCurrent.valueLiteral
	p.advanceToken()

	if !p.expect(tokenAssign) {
		return nil
	}

	result.Value = p.parseExpression(0) // parse the value expression
	if result.Value == nil {
		p.errorf("invalid setting value expression")
		return nil
	}

	if !p.expect(tokenSemicolon) {
		return nil
	}

	return &result
}

func (p *Parser) parseMonitor() *Monitor {
	var result Monitor

	if !p.expectIdentifier("monitor") {
		return nil
	}

	if p.tokenCurrent.kind != tokenStringLiteral {
		p.errorf("expected monitor column name string, got %v", p.tokenCurrent.kind)
		return nil
	}

	result.ColumnName = p.tokenCurrent.valueLiteral

	p.advanceToken()

	if !p.expect(tokenLeftBrace) {
		return nil
	}

	for p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		if p.tokenCurrent.kind == tokenIdentifier && p.tokenCurrent.valueLiteral == "level" {
			r := p.parseRule()
			if r != nil {
				result.Rules = append(result.Rules, r)
			} else {
				// error recovery: skip to next rule or '}'
				p.skipToIdentifierRightBrace("level")
			}
		} else {
			p.errorf("unexpected token inside monitor block: %v (%s)", p.tokenCurrent.kind, p.tokenCurrent.valueLiteral)
			p.skipToIdentifierRightBrace("level")
		}
	}

	if !p.expect(tokenRightBrace) {
		p.errorf("monitor block not properly closed")
		if p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF {
			p.advanceToken()
		}
	}

	return &result
}

func (p *Parser) parseRule() *Rule {
	var result Rule

	if !p.expectIdentifier("level") {
		return nil
	}

	if p.tokenCurrent.kind != tokenNumber {
		p.errorf("expected rule level number, got %v", p.tokenCurrent.kind)
		return nil
	}
	level, err := strconv.Atoi(p.tokenCurrent.valueLiteral)
	if err != nil {
		p.errorf("invalid level number '%s': %v", p.tokenCurrent.valueLiteral, err)
		return nil
	}

	result.Level = level

	p.advanceToken()

	if !p.expectIdentifier("when") {
		return nil
	}

	result.Condition = p.parseExpression(0) // parse the condition expression
	if result.Condition == nil {
		p.errorf("invalid rule condition expression")
		return nil
	}

	if !p.expect(tokenSemicolon) {
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

func parse(input io.Reader) (*AlertConfiguration, []string) {
	buf := make([]byte, 1)
	n, err := input.Read(buf)
	if n == 0 || err == io.EOF {
		return nil, []string{"input is empty"}
	}

	l := newlexer(
		io.MultiReader(bytes.NewReader(buf), input),
	)

	p := NewParser(l)
	p.debug = true // TODO: remove later on.

	programAST := p.parserEntrypoint()

	if l.errorParsing != nil {
		p.errors = append(p.errors, fmt.Sprintf("lexer error: %v", l.errorParsing))
	}

	return programAST, p.errors
}
