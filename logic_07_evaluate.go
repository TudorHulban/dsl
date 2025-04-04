package dslalert

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	goerrors "github.com/TudorHulban/go-errors"
)

func EvaluateCriteria(criteria *Criteria, dataset []string) (EvaluationResults, error) {
	if criteria == nil {
		return nil,
			goerrors.ErrValidation{
				Caller: "EvaluateCriteria",
				Issue: goerrors.ErrNilInput{
					InputName: "criteria",
				},
			}
	}

	if len(dataset) == 0 {
		return nil,
			goerrors.ErrValidation{
				Caller: "EvaluateCriteria",
				Issue: goerrors.ErrNilInput{
					InputName: "dataset",
				},
			}
	}

	// column name | column number
	mapHeaderColumns := make(map[string]int)

	header := strings.Split(dataset[0], ",")

	for ix, nameColumn := range header {
		mapHeaderColumns[nameColumn] = ix
	}

	var results []EvaluationResult
	rowIndex := 1

	// Read data rows
	for rowIndex < len(dataset) {
		record := strings.Split(dataset[rowIndex], ",")
		rowIndex++

		if len(record) != len(mapHeaderColumns) {
			continue
		}

		for _, monitor := range criteria.Monitors {
			colIdx, ok := mapHeaderColumns[monitor.ColumnName]
			if !ok {
				continue // Column not found in header, skip monitor
			}

			if colIdx >= len(record) {
				fmt.Printf(
					"Warning: Row %d is too short for monitored column '%s'\n",
					rowIndex+1,
					monitor.ColumnName,
				)

				continue // Skip if row doesn't have enough columns
			}

			rawValue := record[colIdx]

			// Attempt to convert raw string value to float64 for numeric comparison
			currentValue, err := strconv.ParseFloat(rawValue, 64)
			if err != nil {
				// Skip numeric evaluation if value isn't numeric
				continue
			}

			// Sort rules by level descending
			sort.SliceStable(monitor.Rules, func(i, j int) bool {
				return monitor.Rules[i].Level > monitor.Rules[j].Level
			})

			// Check rules for this monitor (highest level first)
			for _, rule := range monitor.Rules {
				// Evaluate the rule's condition (without settings)
				match, evalErr := evaluateCondition(rule.Condition, currentValue) // Pass only value
				if evalErr != nil {
					fmt.Printf("Warning: Row %d, Criteria '%s', Monitor '%s': Error evaluating condition for level %d: %v\n",
						rowIndex+1, criteria.Name, monitor.ColumnName, rule.Level, evalErr)
					continue // Skip rule if evaluation fails
				}

				if match {
					// Highest level rule matched
					result := EvaluationResult{
						RowIndex:     rowIndex,
						CriteriaName: criteria.Name,
						MonitorName:  monitor.ColumnName,
						RuleLevel:    rule.Level,
						Message: fmt.Sprintf("Row %d: Alert triggered! Criteria='%s', Monitor='%s' (%s), Level=%d (Value=%v)",
							rowIndex+1, criteria.Name, monitor.ColumnName, header[colIdx], rule.Level, currentValue),
					}
					results = append(results, result)
					break // Stop checking lower levels for this monitor on this row
				}
			} // end rule loop
		} // end monitor loop

		rowIndex++
	} // end row loop

	return results, nil
}

func evaluateCondition(expr Expression, contextValue float64) (bool, error) {
	// Call evaluateExpression without settings map
	result, err := evaluateExpression(expr, contextValue)
	if err != nil {
		return false, err
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("condition expression did not evaluate to a boolean, got %T", result)
	}

	return boolResult, nil
}

// Only variable allowed is 'value'.
func evaluateExpression(expr Expression, contextValue float64) (interface{}, error) {
	switch e := expr.(type) {
	case *ExpressionLiteral:
		return e.value, nil // Return literal value

	case *ExpressionVariable:
		if e.name == "value" {
			return contextValue, nil // Substitute the special 'value' variable
		}

		// No other variables are allowed in this simplified version
		return nil,
			fmt.Errorf(
				"undefined variable '%s' (only 'value' is allowed)",
				e.name,
			)

	case *ExpressionBinary:
		// Recursively evaluate left and right sides
		leftVal, err := evaluateExpression(e.LefthandSide, contextValue)
		if err != nil {
			return nil,
				fmt.Errorf(
					"failed to evaluate left side of '%s': %w",
					e.Operator,
					err,
				)
		}

		rightVal, err := evaluateExpression(e.RighthandSide, contextValue)
		if err != nil {
			return nil,
				fmt.Errorf(
					"failed to evaluate right side of '%s': %w",
					e.Operator,
					err,
				)
		}

		// Perform Operation (numeric focus, same as before)
		leftFloat, leftOk := toFloat64(leftVal)
		rightFloat, rightOk := toFloat64(rightVal)

		if isComparisonOperator(e.Operator) {
			if !leftOk || !rightOk {
				return nil,
					fmt.Errorf(
						"cannot compare non-numeric values ('%v' %s '%v')",
						leftVal,
						e.Operator,
						rightVal,
					)
			}

			switch e.Operator {
			case ">":
				return leftFloat > rightFloat, nil
			case ">=":
				return leftFloat >= rightFloat, nil
			case "<":
				return leftFloat < rightFloat, nil
			case "<=":
				return leftFloat <= rightFloat, nil
			case "==":
				return leftFloat == rightFloat, nil
			case "!=":
				return leftFloat != rightFloat, nil

			default:
				return nil, fmt.Errorf("unsupported comparison operator '%s'", e.Operator)
			}
		}

		if isArithmeticOperator(e.Operator) {
			if !leftOk || !rightOk {
				return nil, fmt.Errorf("cannot perform arithmetic on non-numeric values ('%v' %s '%v')", leftVal, e.Operator, rightVal)
			}

			switch e.Operator {
			case "+":
				return leftFloat + rightFloat, nil
			case "-":
				return leftFloat - rightFloat, nil
			case "*":
				return leftFloat * rightFloat, nil
			case "/":
				if rightFloat == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return leftFloat / rightFloat, nil

			default:
				return nil, fmt.Errorf("unsupported arithmetic operator '%s'", e.Operator)
			}
		}

		return nil, fmt.Errorf("unsupported binary operator '%s'", e.Operator)

	default:
		return nil, fmt.Errorf("unsupported expression type %T", expr)
	}
}

// --- Example Usage (Simplified - No Settings) ---
/*
func main() {
	// 1. Construct Criteria AST (Simplified - no settings allowed)
	criteriaExample := &Criteria{
		Name: "value_limits",
		// No Settings field
		Monitors: []*Monitor{
			{
				ColumnName: "order_count",
				Rules: []*Rule{
					{
						Level: 1,
						Condition: &ExpressionBinary{ // value > 100.0
							Op:  ">",
							Lhs: &ExpressionVariable{Name: "value"},
							Rhs: &ExpressionLiteral{Value: 100.0, Raw: "100.0"},
						},
					},
                    {
						Level: 2,
						Condition: &ExpressionBinary{ // value >= 150.0
							Op:  ">=",
							Lhs: &ExpressionVariable{Name: "value"},
                            Rhs: &ExpressionLiteral{Value: 150.0, Raw: "150.0"},
						},
					},
				},
			},
            {
                ColumnName: "amount",
                 Rules: []*Rule{
                    {
                        Level: 1,
                        Condition: &ExpressionBinary{ // value > 500.50
                            Op: ">",
                            Lhs: &ExpressionVariable{Name: "value"},
                            Rhs: &ExpressionLiteral{Value: 500.50, Raw:"500.50"},
                        },
                    },
                 },
            },
		},
	}

	// 2. Provide CSV data
	csvData := `order_count,customer_id,amount
90,cust1,450.00
110,cust2,600.00
155,cust3,300.00
75,cust4,700.75
`

	// 3. Evaluate
	results, err := EvaluateCriteria(criteriaExample, csvData)

	// 4. Process results
	if err != nil {
		fmt.Printf("Error evaluating criteria: %v\n", err)
	} else {
		fmt.Println("Evaluation Results:")
		if len(results) == 0 {
			fmt.Println("  No alerts triggered.")
		}
		for _, res := range results {
			fmt.Printf("  - %s\n", res.Message)
		}
	}
}

// Dummy Expression implementations needed for main() to compile
type ExpressionBinary struct { Op string; Lhs, Rhs Expression }
func (e *ExpressionBinary) interfaceMarker() {}
func (e *ExpressionBinary) String() string { return fmt.Sprintf("(%s %s %s)", e.Lhs.String(), e.Op, e.Rhs.String()) }
type ExpressionLiteral struct { Value interface{}; Raw string }
func (e *ExpressionLiteral) interfaceMarker() {}
func (e *ExpressionLiteral) String() string { return e.Raw }
type ExpressionVariable struct { Name string }
func (e *ExpressionVariable) interfaceMarker() {}
func (e *ExpressionVariable) String() string { return e.Name }
*/
