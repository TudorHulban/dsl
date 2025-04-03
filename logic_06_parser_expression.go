package main

import "strconv"

func (p *Parser) currentPrecedence() int {
	if p.tokenCurrent.kind != tokenOperator {
		return 0
	}

	return p.operatorPrecedence(p.tokenCurrent.valueLiteral)
}

func (p *Parser) operatorPrecedence(op string) int {
	switch op {
	case "*", "/":
		return 5
	case "+", "-":
		return 4
	case ">", "<", ">=", "<=", "==", "!=":
		return 3

	default:
		return 0
	}
}

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
		left = newVariable(p.tokenCurrent.valueLiteral)

		p.advanceToken()

	default:
		p.errorf(
			"unexpected token in expression: %v (%s)",
			p.tokenCurrent.kind,
			p.tokenCurrent.valueLiteral,
		)

		return nil
	}

	for {
		// Stop if next token isn't an operator or at higher precedence
		if p.tokenCurrent.kind != tokenOperator ||
			precedence >= p.currentPrecedence() {
			break
		}

		op := p.tokenCurrent.valueLiteral
		opPrec := p.operatorPrecedence(op)
		p.advanceToken()

		right := p.parseExpression(opPrec)
		if right == nil {
			return nil
		}

		left = &ExpressionBinary{
			LefthandSide:  left,
			Operator:      op,
			RighthandSide: right,
		}
	}

	// look ahead for a binary operator (super simplified)
	// if p.tokenCurrent.kind == tokenOperator {
	// 	op := p.tokenCurrent.valueLiteral

	// 	p.advanceToken()

	// 	right := p.parseExpression(0) // recursive call (doesn't handle precedence)
	// 	if right == nil {
	// 		p.errorf(
	// 			"missing right hand side for operator %s",
	// 			op,
	// 		)

	// 		return nil
	// 	}

	// 	return newbinaryexpr(left, op, right)
	// }

	return left // return just the literal or variable if no operator follows
}
