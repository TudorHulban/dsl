package dslalert

type expressionLiteral struct {
	value any    // use type assertion later (e.g., float64, string)
	raw   string // original raw string representation
}

func (e *expressionLiteral) interfaceMarker() {}

func (e *expressionLiteral) String() string {
	return e.raw
}

func newliteral(value any, raw string) *expressionLiteral {
	return &expressionLiteral{
		value: value,
		raw:   raw,
	}
}
