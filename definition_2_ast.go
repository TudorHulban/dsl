package dslalert

import "fmt"

type Expression interface {
	interfaceMarker()
	String() string
}

var _ Expression = &ExpressionBinary{}
var _ Expression = &ExpressionLiteral{}
var _ Expression = &ExpressionVariable{}

type Rule struct {
	Level     int
	Condition Expression // the 'when' condition expression
}

type Monitor struct {
	ColumnName string
	Rules      []*Rule
}

type Criteria struct {
	Name     string
	Monitors []*Monitor
}

type AlertConfiguration struct {
	Criterias []*Criteria
}

type EvaluationResult struct {
	RowIndex     int // 0-based index of the data row (1st data row is index 0)
	CriteriaName string
	MonitorName  string
	RuleLevel    int
	Message      string // Pre-formatted alert message
}

func (e EvaluationResult) String() string {
	return fmt.Sprintf(
		"RowIndex: %d, CriteriaName: %s, MonitorName: %s, RuleLevel: %d, Message: %s",

		e.RowIndex,
		e.CriteriaName,
		e.MonitorName,
		e.RuleLevel,
		e.Message,
	)
}

type EvaluationResults []EvaluationResult

func (results EvaluationResults) String() string {
	var result string

	for _, value := range results {
		result = result + value.String() + "\n"
	}

	return result
}
