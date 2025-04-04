package dslalert

import "fmt"

type expressionBinary struct {
	Operator      string // (e.g., ">=", "<", "+", "==")
	LefthandSide  expression
	RighthandSide expression
}

func (e *expressionBinary) interfaceMarker() {}

func (e *expressionBinary) String() string {
	return fmt.Sprintf(
		"(%s %s %s)",

		e.LefthandSide.String(),
		e.Operator,
		e.RighthandSide.String(),
	)
}
