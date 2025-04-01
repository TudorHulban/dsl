package main

import "fmt"

type exprBinary struct {
	operator string     // operator (e.g., ">=", "<", "+", "==")
	lhs      expression // left hand side
	rhs      expression // right hand side
}

func (e *exprBinary) exprNode() {}

func (e *exprBinary) string() string {
	return fmt.Sprintf(
		"(%s %s %s)",

		e.lhs.string(),
		e.operator,
		e.rhs.string(),
	)
}

func newbinaryexpr(lhs expression, op string, rhs expression) *exprBinary {
	return &exprBinary{lhs: lhs, operator: op, rhs: rhs}
}
