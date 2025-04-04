package dslalert

import "fmt"

type EvaluationResult struct {
	ValueCurrent any

	CriteriaName string
	MonitorName  string
	Row          string

	RowIndex  int
	RuleLevel int
}

func (e EvaluationResult) String() string {
	return fmt.Sprintf(
		"Row %d: %s --> Alert triggered! Level=%d (Value=%v)",
		e.RowIndex,
		e.Row,
		e.RuleLevel,
		e.ValueCurrent,
	)
}

type EvaluationResults []EvaluationResult

func (results EvaluationResults) LevelMaximum() int {
	var result int

	for _, r := range results {
		if r.RuleLevel > result {
			result = r.RuleLevel
		}
	}

	return result
}

func (results EvaluationResults) Rows() []string {
	result := make([]string, len(results), len(results))

	for ix, r := range results {
		result[ix] = r.Row
	}

	return result
}

func (results EvaluationResults) Message() string {
	if len(results) == 0 {
		return ""
	}

	return fmt.Sprintf(
		"Level maximum for criteria %s (monitor %s): %d",

		results[0].CriteriaName,
		results[0].MonitorName,
		results.LevelMaximum(),
	)
}

func (results EvaluationResults) String() string {
	var result string

	for _, value := range results {
		result = result + value.String() + "\n"
	}

	return result
}
