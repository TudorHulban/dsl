package main

import (
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

	p.advanceToken()
	p.advanceToken()

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
func (p *Parser) expect(t tokentype) bool {
	if p.tokenCurrent.typ == t {
		p.advanceToken()
		return true
	}

	p.errorf("expected token %v, got %v (%s)", t, p.tokenCurrent.typ, p.tokenCurrent.lit)

	return false
}

// expectIdentifier checks if the current token is an identifier with specific text.
func (p *Parser) expectIdentifier(ident string) bool {
	if p.tokenCurrent.typ == tokenident && p.tokenCurrent.lit == ident {
		p.advanceToken()
		return true
	}

	p.errorf("expected identifier '%s', got %v (%s)", ident, p.tokenCurrent.typ, p.tokenCurrent.lit)

	return false
}

// --- parsing functions (recursive descent) ---

// parseProgram is the entry point.
func (p *Parser) parseProgram() *program {
	prog := &program{}

	for p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
		if p.tokenCurrent.typ == tokenident && p.tokenCurrent.lit == "dataset" {
			ds := p.parseDataset()
			if ds != nil {
				prog.datasets = append(prog.datasets, ds)
			} else {
				// basic error recovery: skip until next potential dataset or eof
				p.skiptoident("dataset")
			}
		} else {
			p.errorf("unexpected token at program root: %v (%s)", p.tokenCurrent.typ, p.tokenCurrent.lit)
			// basic error recovery
			p.skiptoident("dataset")
		}
	}

	return prog
}

func (p *Parser) parseDataset() *dataset {
	if !p.expectIdentifier("dataset") {
		return nil
	}

	// Expect and capture the dataset name
	if p.tokenCurrent.typ != tokenstring {
		p.errorf("expected dataset name string, got %v", p.tokenCurrent.typ)
		return nil
	}
	ds := &dataset{name: p.tokenCurrent.lit}
	p.advanceToken()

	// Expect opening brace
	if !p.expect(tokenlbrace) {
		return nil
	}

	// Temporary: Just consume tokens until closing brace
	for p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof {
		p.advanceToken()
	}

	// Expect closing brace
	if !p.expect(tokenrbrace) {
		return nil
	}

	return ds
}

// func (p *Parser) parseDataset() *dataset {
// 	ds := &dataset{}
// 	if !p.expectIdentifier("dataset") {
// 		return nil
// 	}

// 	if p.tokenCurrent.typ != tokenstring {
// 		p.errorf("expected dataset name string, got %v", p.tokenCurrent.typ)
// 		return nil
// 	}
// 	ds.name = p.tokenCurrent.lit
// 	p.advanceToken()

// 	if !p.expect(tokenlbrace) {
// 		return nil
// 	}

// 	for p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
// 		if p.tokenCurrent.typ == tokenident && p.tokenCurrent.lit == "criteria" {
// 			crit := p.parseCriteria()
// 			if crit != nil {
// 				ds.criteria = append(ds.criteria, crit)
// 			} else {
// 				// error recovery: skip to next criteria or '}'
// 				p.skiptoidentorbrace("criteria")
// 			}
// 		} else {
// 			p.errorf("unexpected token inside dataset block: %v (%s)", p.tokenCurrent.typ, p.tokenCurrent.lit)
// 			p.skiptoidentorbrace("criteria")
// 		}
// 	}

// 	if !p.expect(tokenrbrace) {
// 		p.errorf("dataset block not properly closed") // error already added by expect
// 		// attempt to recover by advancing if stuck
// 		if p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof {
// 			p.advanceToken()
// 		}
// 	}

// 	return ds
// }

func (p *Parser) parseCriteria() *criteria {
	crit := &criteria{}
	if !p.expectIdentifier("criteria") {
		return nil
	}

	if p.tokenCurrent.typ != tokenstring {
		p.errorf("expected criteria name string, got %v", p.tokenCurrent.typ)
		return nil
	}
	crit.name = p.tokenCurrent.lit
	p.advanceToken()

	if !p.expect(tokenlbrace) {
		return nil
	}

	for p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
		switch {
		case p.tokenCurrent.typ == tokenident && (p.tokenCurrent.lit == "baseline" || p.tokenCurrent.lit == "increment"):
			sett := p.parseSetting()
			if sett != nil {
				crit.settings = append(crit.settings, sett)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skiptoidentorbrace("baseline", "increment", "monitor")
			}

		case p.tokenCurrent.typ == tokenident && p.tokenCurrent.lit == "monitor":
			mon := p.parseMonitor()
			if mon != nil {
				crit.monitors = append(crit.monitors, mon)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skiptoidentorbrace("baseline", "increment", "monitor")
			}

		default:
			p.errorf("unexpected token inside criteria block: %v (%s)", p.tokenCurrent.typ, p.tokenCurrent.lit)
			p.skiptoidentorbrace("baseline", "increment", "monitor")
		}
	}

	if !p.expect(tokenrbrace) {
		p.errorf("criteria block not properly closed")
		if p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof {
			p.advanceToken()
		}
	}

	return crit
}

func (p *Parser) parseSetting() *setting {
	sett := &setting{}
	sett.kind = p.tokenCurrent.lit // "baseline" or "increment"
	p.advanceToken()

	if p.tokenCurrent.typ != tokenident {
		p.errorf("expected setting name identifier, got %v", p.tokenCurrent.typ)
		return nil
	}
	sett.name = p.tokenCurrent.lit
	p.advanceToken()

	if !p.expect(tokenassign) {
		return nil
	}

	sett.value = p.parseExpression(0) // parse the value expression
	if sett.value == nil {
		p.errorf("invalid setting value expression")
		return nil
	}

	if !p.expect(tokensemicolon) {
		return nil
	}
	return sett
}

func (p *Parser) parseMonitor() *monitor {
	mon := &monitor{}
	if !p.expectIdentifier("monitor") {
		return nil
	}

	if p.tokenCurrent.typ != tokenstring {
		p.errorf("expected monitor column name string, got %v", p.tokenCurrent.typ)
		return nil
	}
	mon.columnname = p.tokenCurrent.lit
	p.advanceToken()

	if !p.expect(tokenlbrace) {
		return nil
	}

	for p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
		if p.tokenCurrent.typ == tokenident && p.tokenCurrent.lit == "level" {
			r := p.parseRule()
			if r != nil {
				mon.rules = append(mon.rules, r)
			} else {
				// error recovery: skip to next rule or '}'
				p.skiptoidentorbrace("level")
			}
		} else {
			p.errorf("unexpected token inside monitor block: %v (%s)", p.tokenCurrent.typ, p.tokenCurrent.lit)
			p.skiptoidentorbrace("level")
		}
	}

	if !p.expect(tokenrbrace) {
		p.errorf("monitor block not properly closed")
		if p.tokenCurrent.typ != tokenrbrace && p.tokenCurrent.typ != tokeneof {
			p.advanceToken()
		}
	}
	return mon
}

func (p *Parser) parseRule() *rule {
	r := &rule{}
	if !p.expectIdentifier("level") {
		return nil
	}

	if p.tokenCurrent.typ != tokennumber {
		p.errorf("expected rule level number, got %v", p.tokenCurrent.typ)
		return nil
	}
	level, err := strconv.Atoi(p.tokenCurrent.lit)
	if err != nil {
		p.errorf("invalid level number '%s': %v", p.tokenCurrent.lit, err)
		return nil
	}
	r.level = level
	p.advanceToken()

	if !p.expectIdentifier("when") {
		return nil
	}

	r.condition = p.parseExpression(0) // parse the condition expression
	if r.condition == nil {
		p.errorf("invalid rule condition expression")
		return nil
	}

	if !p.expect(tokensemicolon) {
		return nil
	}
	return r
}

// parseExpression - simplified placeholder for expression parsing
// a real implementation needs operator precedence (e.g., Pratt parsing or shunting-yard)
func (p *Parser) parseExpression(precedence int) expression {
	// very basic: handles literal or variable, optionally followed by operator and another term
	// does not handle precedence or parentheses correctly!
	var left expression

	switch p.tokenCurrent.typ {
	case tokennumber:
		// try parsing as float first
		fval, errf := strconv.ParseFloat(p.tokenCurrent.lit, 64)
		if errf == nil {
			left = newliteral(fval, p.tokenCurrent.lit)
		} else {
			// try parsing as int
			ival, erri := strconv.Atoi(p.tokenCurrent.lit)
			if erri == nil {
				left = newliteral(ival, p.tokenCurrent.lit)
			} else {
				p.errorf("invalid number literal: %s", p.tokenCurrent.lit)
				return nil
			}
		}
		p.advanceToken()

	case tokenident:
		left = newvariable(p.tokenCurrent.lit)
		p.advanceToken()

	default:
		p.errorf("unexpected token in expression: %v (%s)", p.tokenCurrent.typ, p.tokenCurrent.lit)
		return nil
	}

	// look ahead for a binary operator (super simplified)
	if p.tokenCurrent.typ == tokenoperator {
		op := p.tokenCurrent.lit
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

func (p *Parser) skipto(types ...tokentype) {
	for p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
		for _, t := range types {
			if p.tokenCurrent.typ == t {
				return // found one of the target types
			}
		}
		p.advanceToken()
	}
}

func (p *Parser) skiptoident(idents ...string) {
	for p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
		if p.tokenCurrent.typ == tokenident {
			for _, id := range idents {
				if p.tokenCurrent.lit == id {
					return // found one of the target idents
				}
			}
		}

		p.advanceToken()
	}
}

func (p *Parser) skiptoidentorbrace(idents ...string) {
	for p.tokenCurrent.typ != tokeneof && p.tokenCurrent.typ != tokenerror {
		if p.tokenCurrent.typ == tokenrbrace { // stop at closing brace
			return
		}
		if p.tokenCurrent.typ == tokenident {
			for _, id := range idents {
				if p.tokenCurrent.lit == id {
					return // found one of the target idents
				}
			}
		}
		p.advanceToken()
	}
}

func (p *Parser) currentTokenIs(t tokentype) bool {
	return p.tokenCurrent.typ == t
}

func (p *Parser) advance() {
	p.tokenCurrent = p.tokenNext

	p.tokenNext = p.lex.nextToken()
}

// --- Main parse function ---
func parse(input io.Reader) (*program, []string) {
	l := newlexer(input)
	p := NewParser(l)
	programast := p.parseProgram()
	// check for lexer errors accumulated during parsing
	if l.err != nil {
		p.errors = append(p.errors, fmt.Sprintf("lexer error: %s", l.err.Error()))
	}
	return programast, p.errors
}
