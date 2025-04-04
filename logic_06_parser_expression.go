package dslalert

import "strconv"

func (p *parser) currentPrecedence() int {
	if p.tokenCurrent.kind != tokenOperator {
		return 0
	}

	return p.operatorPrecedence(p.tokenCurrent.valueLiteral)
}

func (p *parser) operatorPrecedence(op string) int {
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

func (p *parser) parseExpression(precedence int) expression {
	var left expression

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
		// Stop if next token is not an operator or at higher precedence
		if p.tokenCurrent.kind != tokenOperator ||
			precedence >= p.currentPrecedence() {
			break
		}

		currentOperator := p.tokenCurrent.valueLiteral
		opPrec := p.operatorPrecedence(currentOperator)

		p.advanceToken()

		right := p.parseExpression(opPrec)
		if right == nil {
			return nil
		}

		left = &expressionBinary{
			LefthandSide:  left,
			Operator:      currentOperator,
			RighthandSide: right,
		}
	}

	return left // return just the literal or variable if no operator follows
}
