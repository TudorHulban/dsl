package dslalert

type expression interface {
	interfaceMarker() // added as other types could implement stringers.
	string() string
}

var _ expression = &expressionBinary{}
var _ expression = &expressionLiteral{}
var _ expression = &expressionVariable{}

type rule struct {
	Level     int
	Condition expression // the 'when' condition expression
}

type monitor struct {
	ColumnName string
	Rules      []*rule
}

type criteria struct {
	Name     string
	Monitors []*monitor
}

type AlertConfiguration struct {
	Criterias []*criteria
}
