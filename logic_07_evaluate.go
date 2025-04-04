package dslalert

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	goerrors "github.com/TudorHulban/go-errors"
)

func evaluateExpression(expr Expression, contextValue float64) (any, error) {
	switch expressionType := expr.(type) {
	case *ExpressionLiteral:
		return expressionType.value, nil // Return literal value

	case *ExpressionVariable:
		if expressionType.name == "value" {
			return contextValue, nil // Substitute the special 'value' variable
		}

		return nil,
			fmt.Errorf(
				"undefined variable '%s' (only 'value' is allowed)",
				expressionType.name,
			)

	case *ExpressionBinary:
		// Recursively evaluate left and right sides
		valueLeft, errEvaluateLeft := evaluateExpression(expressionType.LefthandSide, contextValue)
		if errEvaluateLeft != nil {
			return nil,
				fmt.Errorf(
					"failed to evaluate left side of '%s': %w",
					expressionType.Operator,
					errEvaluateLeft,
				)
		}

		valueRight, errEvaluateRight := evaluateExpression(expressionType.RighthandSide, contextValue)
		if errEvaluateRight != nil {
			return nil,
				fmt.Errorf(
					"failed to evaluate right side of '%s': %w",
					expressionType.Operator,
					errEvaluateRight,
				)
		}

		// Perform Operation (numeric focus, same as before)
		floatLeft, leftOk := toFloat64(valueLeft)
		floatRight, rightOk := toFloat64(valueRight)

		if isComparisonOperator(expressionType.Operator) {
			if !leftOk || !rightOk {
				return nil,
					fmt.Errorf(
						"cannot compare non-numeric values ('%v' %s '%v')",
						valueLeft,
						expressionType.Operator,
						valueRight,
					)
			}

			switch expressionType.Operator {
			case ">":
				return floatLeft > floatRight, nil
			case ">=":
				return floatLeft >= floatRight, nil
			case "<":
				return floatLeft < floatRight, nil
			case "<=":
				return floatLeft <= floatRight, nil
			case "==":
				return floatLeft == floatRight, nil
			case "!=":
				return floatLeft != floatRight, nil

			default:
				return nil,
					fmt.Errorf(
						"unsupported comparison operator '%s'",
						expressionType.Operator,
					)
			}
		}

		if isArithmeticOperator(expressionType.Operator) {
			if !leftOk || !rightOk {
				return nil,
					fmt.Errorf(
						"cannot perform arithmetic on non-numeric values ('%v' %s '%v')",
						valueLeft,
						expressionType.Operator,
						valueRight,
					)
			}

			switch expressionType.Operator {
			case "+":
				return floatLeft + floatRight, nil
			case "-":
				return floatLeft - floatRight, nil
			case "*":
				return floatLeft * floatRight, nil
			case "/":
				if floatRight == 0 {
					return nil, fmt.Errorf("division by zero")
				}

				return floatLeft / floatRight, nil

			default:
				return nil,
					fmt.Errorf(
						"unsupported arithmetic operator '%s'",
						expressionType.Operator,
					)
			}
		}

		return nil,
			fmt.Errorf(
				"unsupported binary operator '%s'",
				expressionType.Operator,
			)

	default:
		return nil,
			fmt.Errorf(
				"unsupported expression type %T",
				expr,
			)
	}
}

func evaluateCondition(expr Expression, contextValue float64) (bool, error) {
	result, errEvaluate := evaluateExpression(expr, contextValue)
	if errEvaluate != nil {
		return false,
			errEvaluate
	}

	resultBoolean, couldCast := result.(bool)
	if !couldCast {
		return false,
			fmt.Errorf(
				"condition expression did not evaluate to a boolean, got %T",
				result,
			)
	}

	return resultBoolean, nil
}

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
	mapHeaderColumns := make(map[string]int, 0)

	header := strings.Split(dataset[0], ",")

	for ix, nameColumn := range header {
		mapHeaderColumns[nameColumn] = ix
	}

	var results []EvaluationResult
	rowIndex := 1

	for rowIndex < len(dataset) {
		record := strings.Split(dataset[rowIndex], ",")

		if len(record) != len(mapHeaderColumns) {
			fmt.Printf(
				"Warning: Number fields row %d is different than header '%d'\n",
				rowIndex,
				len(mapHeaderColumns),
			)

			continue
		}

		for _, monitor := range criteria.Monitors {
			columnIx, exists := mapHeaderColumns[monitor.ColumnName]
			if !exists {
				continue
			}

			valueRaw := record[columnIx]

			// for numeric comparison
			valueCurrent, err := strconv.ParseFloat(valueRaw, 64)
			if err != nil {
				continue
			}

			// Sort rules by level descending
			sort.SliceStable(
				monitor.Rules,
				func(i, j int) bool {
					return monitor.Rules[i].Level > monitor.Rules[j].Level
				},
			)

			for _, rule := range monitor.Rules {
				match, errEvaluate := evaluateCondition(rule.Condition, valueCurrent)
				if errEvaluate != nil {
					fmt.Printf(
						"Warning: Row %d, Criteria '%s', Monitor '%s': Error evaluating condition for level %d: %v\n",
						rowIndex,
						criteria.Name,
						monitor.ColumnName,
						rule.Level,
						errEvaluate,
					)

					continue
				}

				if match {
					result := EvaluationResult{
						CriteriaName: criteria.Name,
						MonitorName:  monitor.ColumnName,
						Row:          dataset[rowIndex],

						RowIndex:  rowIndex,
						RuleLevel: rule.Level,

						ValueCurrent: valueCurrent,
					}

					results = append(results, result)

					break
				}
			}
		}

		rowIndex++
	}

	return results,
		nil
}
