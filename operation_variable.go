package main

// exprVariable represents a variable name (e.g., 'value', 'setting_name', baseline names).
type exprVariable struct {
	name string
}

func (e *exprVariable) exprNode() {}

func (e *exprVariable) string() string { return e.name }

func newvariable(name string) *exprVariable {
	return &exprVariable{name: name}
}
