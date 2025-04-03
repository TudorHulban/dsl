package main

// ExpressionVariable represents a variable name (e.g., 'value', 'setting_name', baseline names).
type ExpressionVariable struct {
	name string
}

func (e *ExpressionVariable) exprNode() {}

func (e *ExpressionVariable) String() string {
	return e.name
}

func newvariable(name string) *ExpressionVariable {
	return &ExpressionVariable{
		name: name,
	}
}
