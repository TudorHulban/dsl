package main

import (
	"fmt"
	"io"
	"strconv"
)

// parser holds the state of the parsing process.
type parser struct {
	l       *lexer
	curtok  token // current token
	peektok token // next token
	errors  []string
}

func newparser(l *lexer) *parser {
	p := &parser{l: l}
	// read two tokens to initialize curtok and peektok
	p.advancetok()
	p.advancetok()

	return p
}

// advancetok moves to the next token.
func (p *parser) advancetok() {
	p.curtok = p.peektok
	p.peektok = p.l.nexttoken()
}

// errorf records a parsing error.
func (p *parser) errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf("parse error at %s: %s", p.curtok.pos, fmt.Sprintf(format, args...))
	p.errors = append(p.errors, msg)
}

// expect checks if the current token matches the expected type.
func (p *parser) expect(t tokentype) bool {
	if p.curtok.typ == t {
		p.advancetok()
		return true
	}

	p.errorf("expected token %v, got %v (%s)", t, p.curtok.typ, p.curtok.lit)

	return false
}

// expectIdentifier checks if the current token is an identifier with specific text.
func (p *parser) expectIdentifier(ident string) bool {
	if p.curtok.typ == tokenident && p.curtok.lit == ident {
		p.advancetok()
		return true
	}

	p.errorf("expected identifier '%s', got %v (%s)", ident, p.curtok.typ, p.curtok.lit)

	return false
}

// --- parsing functions (recursive descent) ---

// parseProgram is the entry point.
func (p *parser) parseProgram() *program {
	prog := &program{}

	for p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		if p.curtok.typ == tokenident && p.curtok.lit == "dataset" {
			ds := p.parseDataset()
			if ds != nil {
				prog.datasets = append(prog.datasets, ds)
			} else {
				// basic error recovery: skip until next potential dataset or eof
				p.skiptoident("dataset")
			}
		} else {
			p.errorf("unexpected token at program root: %v (%s)", p.curtok.typ, p.curtok.lit)
			// basic error recovery
			p.skiptoident("dataset")
		}
	}

	return prog
}

func (p *parser) parseDataset() *dataset {
	ds := &dataset{}
	if !p.expectIdentifier("dataset") {
		return nil
	}

	if p.curtok.typ != tokenstring {
		p.errorf("expected dataset name string, got %v", p.curtok.typ)
		return nil
	}
	ds.name = p.curtok.lit
	p.advancetok()

	if !p.expect(tokenlbrace) {
		return nil
	}

	for p.curtok.typ != tokenrbrace && p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		if p.curtok.typ == tokenident && p.curtok.lit == "criteria" {
			crit := p.parseCriteria()
			if crit != nil {
				ds.criteria = append(ds.criteria, crit)
			} else {
				// error recovery: skip to next criteria or '}'
				p.skiptoidentorbrace("criteria")
			}
		} else {
			p.errorf("unexpected token inside dataset block: %v (%s)", p.curtok.typ, p.curtok.lit)
			p.skiptoidentorbrace("criteria")
		}
	}

	if !p.expect(tokenrbrace) {
		p.errorf("dataset block not properly closed") // error already added by expect
		// attempt to recover by advancing if stuck
		if p.curtok.typ != tokenrbrace && p.curtok.typ != tokeneof {
			p.advancetok()
		}
	}

	return ds
}

func (p *parser) parseCriteria() *criteria {
	crit := &criteria{}
	if !p.expectIdentifier("criteria") {
		return nil
	}

	if p.curtok.typ != tokenstring {
		p.errorf("expected criteria name string, got %v", p.curtok.typ)
		return nil
	}
	crit.name = p.curtok.lit
	p.advancetok()

	if !p.expect(tokenlbrace) {
		return nil
	}

	for p.curtok.typ != tokenrbrace && p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		switch {
		case p.curtok.typ == tokenident && (p.curtok.lit == "baseline" || p.curtok.lit == "increment"):
			sett := p.parsesetting()
			if sett != nil {
				crit.settings = append(crit.settings, sett)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skiptoidentorbrace("baseline", "increment", "monitor")
			}

		case p.curtok.typ == tokenident && p.curtok.lit == "monitor":
			mon := p.parsemonitor()
			if mon != nil {
				crit.monitors = append(crit.monitors, mon)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skiptoidentorbrace("baseline", "increment", "monitor")
			}

		default:
			p.errorf("unexpected token inside criteria block: %v (%s)", p.curtok.typ, p.curtok.lit)
			p.skiptoidentorbrace("baseline", "increment", "monitor")
		}
	}

	if !p.expect(tokenrbrace) {
		p.errorf("criteria block not properly closed")
		if p.curtok.typ != tokenrbrace && p.curtok.typ != tokeneof {
			p.advancetok()
		}
	}

	return crit
}

func (p *parser) parsesetting() *setting {
	sett := &setting{}
	sett.kind = p.curtok.lit // "baseline" or "increment"
	p.advancetok()

	if p.curtok.typ != tokenident {
		p.errorf("expected setting name identifier, got %v", p.curtok.typ)
		return nil
	}
	sett.name = p.curtok.lit
	p.advancetok()

	if !p.expect(tokenassign) {
		return nil
	}

	sett.value = p.parseexpression(0) // parse the value expression
	if sett.value == nil {
		p.errorf("invalid setting value expression")
		return nil
	}

	if !p.expect(tokensemicolon) {
		return nil
	}
	return sett
}

func (p *parser) parsemonitor() *monitor {
	mon := &monitor{}
	if !p.expectIdentifier("monitor") {
		return nil
	}

	if p.curtok.typ != tokenstring {
		p.errorf("expected monitor column name string, got %v", p.curtok.typ)
		return nil
	}
	mon.columnname = p.curtok.lit
	p.advancetok()

	if !p.expect(tokenlbrace) {
		return nil
	}

	for p.curtok.typ != tokenrbrace && p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		if p.curtok.typ == tokenident && p.curtok.lit == "level" {
			r := p.parserule()
			if r != nil {
				mon.rules = append(mon.rules, r)
			} else {
				// error recovery: skip to next rule or '}'
				p.skiptoidentorbrace("level")
			}
		} else {
			p.errorf("unexpected token inside monitor block: %v (%s)", p.curtok.typ, p.curtok.lit)
			p.skiptoidentorbrace("level")
		}
	}

	if !p.expect(tokenrbrace) {
		p.errorf("monitor block not properly closed")
		if p.curtok.typ != tokenrbrace && p.curtok.typ != tokeneof {
			p.advancetok()
		}
	}
	return mon
}

func (p *parser) parserule() *rule {
	r := &rule{}
	if !p.expectIdentifier("level") {
		return nil
	}

	if p.curtok.typ != tokennumber {
		p.errorf("expected rule level number, got %v", p.curtok.typ)
		return nil
	}
	level, err := strconv.Atoi(p.curtok.lit)
	if err != nil {
		p.errorf("invalid level number '%s': %v", p.curtok.lit, err)
		return nil
	}
	r.level = level
	p.advancetok()

	if !p.expectIdentifier("when") {
		return nil
	}

	r.condition = p.parseexpression(0) // parse the condition expression
	if r.condition == nil {
		p.errorf("invalid rule condition expression")
		return nil
	}

	if !p.expect(tokensemicolon) {
		return nil
	}
	return r
}

// parseexpression - simplified placeholder for expression parsing
// a real implementation needs operator precedence (e.g., Pratt parsing or shunting-yard)
func (p *parser) parseexpression(precedence int) expression {
	// very basic: handles literal or variable, optionally followed by operator and another term
	// does not handle precedence or parentheses correctly!
	var left expression

	switch p.curtok.typ {
	case tokennumber:
		// try parsing as float first
		fval, errf := strconv.ParseFloat(p.curtok.lit, 64)
		if errf == nil {
			left = newliteral(fval, p.curtok.lit)
		} else {
			// try parsing as int
			ival, erri := strconv.Atoi(p.curtok.lit)
			if erri == nil {
				left = newliteral(ival, p.curtok.lit)
			} else {
				p.errorf("invalid number literal: %s", p.curtok.lit)
				return nil
			}
		}
		p.advancetok()
	case tokenident:
		left = newvariable(p.curtok.lit)
		p.advancetok()
	default:
		p.errorf("unexpected token in expression: %v (%s)", p.curtok.typ, p.curtok.lit)
		return nil
	}

	// look ahead for a binary operator (super simplified)
	if p.curtok.typ == tokenoperator {
		op := p.curtok.lit
		p.advancetok()
		right := p.parseexpression(0) // recursive call (doesn't handle precedence)
		if right == nil {
			p.errorf("missing right hand side for operator %s", op)
			return nil
		}
		return newbinaryexpr(left, op, right)
	}

	return left // return just the literal or variable if no operator follows
}

// --- basic error recovery helpers (very naive) ---

func (p *parser) skipto(types ...tokentype) {
	for p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		for _, t := range types {
			if p.curtok.typ == t {
				return // found one of the target types
			}
		}
		p.advancetok()
	}
}

func (p *parser) skiptoident(idents ...string) {
	for p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		if p.curtok.typ == tokenident {
			for _, id := range idents {
				if p.curtok.lit == id {
					return // found one of the target idents
				}
			}
		}
		p.advancetok()
	}
}

func (p *parser) skiptoidentorbrace(idents ...string) {
	for p.curtok.typ != tokeneof && p.curtok.typ != tokenerror {
		if p.curtok.typ == tokenrbrace { // stop at closing brace
			return
		}
		if p.curtok.typ == tokenident {
			for _, id := range idents {
				if p.curtok.lit == id {
					return // found one of the target idents
				}
			}
		}
		p.advancetok()
	}
}

// --- Main parse function ---
func parse(input io.Reader) (*program, []string) {
	l := newlexer(input)
	p := newparser(l)
	programast := p.parseProgram()
	// check for lexer errors accumulated during parsing
	if l.err != nil {
		p.errors = append(p.errors, fmt.Sprintf("lexer error: %s", l.err.Error()))
	}
	return programast, p.errors
}
