package dslalert

import "fmt"

type expressionBinary struct {
	Operator      string // (e.g., ">=", "<", "+", "==")
	LefthandSide  expression
	RighthandSide expression
}

func (e *expressionBinary) interfaceMarker() {}

func (e *expressionBinary) string() string {
	return fmt.Sprintf(
		"(%s %s %s)",

		e.LefthandSide.string(),
		e.Operator,
		e.RighthandSide.string(),
	)
}
