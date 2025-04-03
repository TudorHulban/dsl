package main

import "strconv"

func (p *Parser) parseExpression(precedence int) Expression {
	var left Expression

	switch p.tokenCurrent.kind {
	case tokenNumber:
		valueFloat, errFloat := strconv.ParseFloat(p.tokenCurrent.valueLiteral, 64)
		if errFloat == nil {
			left = newliteral(
				valueFloat,
				p.tokenCurrent.valueLiteral,
			)
		} else {
			valueInteger, errInteger := strconv.Atoi(p.tokenCurrent.valueLiteral)
			if errInteger == nil {
				left = newliteral(
					valueInteger,
					p.tokenCurrent.valueLiteral,
				)
			} else {
				p.errorf(
					"invalid number literal: %s",
					p.tokenCurrent.valueLiteral,
				)

				return nil
			}
		}

		p.advanceToken()

	case tokenIdentifier:
		left = newvariable(p.tokenCurrent.valueLiteral)

		p.advanceToken()

	default:
		p.errorf(
			"unexpected token in expression: %v (%s)",
			p.tokenCurrent.kind,
			p.tokenCurrent.valueLiteral,
		)

		return nil
	}

	// look ahead for a binary operator (super simplified)
	if p.tokenCurrent.kind == tokenOperator {
		op := p.tokenCurrent.valueLiteral

		p.advanceToken()

		right := p.parseExpression(0) // recursive call (doesn't handle precedence)
		if right == nil {
			p.errorf(
				"missing right hand side for operator %s",
				op,
			)

			return nil
		}

		return newbinaryexpr(left, op, right)
	}

	return left // return just the literal or variable if no operator follows
}
