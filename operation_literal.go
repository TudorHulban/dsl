package main

type exprLiteral struct {
	value any    // use type assertion later (e.g., float64, string)
	raw   string // original raw string representation
}

func (e *exprLiteral) exprNode() {}

func (e *exprLiteral) string() string {
	return e.raw
}

func newliteral(value interface{}, raw string) *exprLiteral {
	return &exprLiteral{value: value, raw: raw}
}
