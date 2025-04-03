package main

type ExpressionLiteral struct {
	value any    // use type assertion later (e.g., float64, string)
	raw   string // original raw string representation
}

func (e *ExpressionLiteral) exprNode() {}

func (e *ExpressionLiteral) String() string {
	return e.raw
}

func newliteral(value any, raw string) *ExpressionLiteral {
	return &ExpressionLiteral{
		value: value,
		raw:   raw,
	}
}
