package main

import "fmt"

type ExpressionBinary struct {
	Operator      string // (e.g., ">=", "<", "+", "==")
	LefthandSide  Expression
	RighthandSide Expression
}

func (e *ExpressionBinary) exprNode() {}

func (e *ExpressionBinary) String() string {
	return fmt.Sprintf(
		"(%s %s %s)",

		e.LefthandSide.String(),
		e.Operator,
		e.RighthandSide.String(),
	)
}

func newbinaryexpr(lhs Expression, op string, rhs Expression) *ExpressionBinary {
	return &ExpressionBinary{
		LefthandSide:  lhs,
		Operator:      op,
		RighthandSide: rhs,
	}
}
