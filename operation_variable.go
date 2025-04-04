package dslalert

// expressionVariable represents a variable name (e.g., 'value').
type expressionVariable struct {
	name string
}

func (e *expressionVariable) interfaceMarker() {}

func (e *expressionVariable) String() string {
	return e.name
}

func newVariable(name string) *expressionVariable {
	return &expressionVariable{
		name: name,
	}
}
