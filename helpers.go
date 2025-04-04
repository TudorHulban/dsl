package dslalert

import (
	"fmt"
	"runtime"
	"strconv"
)

// Use as defer traceExit().
func traceExit() {
	pc, _, line, ok := runtime.Caller(1) // Get the caller of this function
	if ok {
		fmt.Printf(
			"exiting function %s at line %d.\n",

			runtime.FuncForPC(pc).Name(),
			line,
		)
	}
}

func isComparisonOperator(operator string) bool {
	switch operator {
	case ">", ">=", "<", "<=", "==", "!=":
		return true

	default:
		return false
	}
}

func isArithmeticOperator(operator string) bool {
	switch operator {
	case "+", "-", "*", "/":
		return true

	default:
		return false
	}
}

func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}

		return 0, false

	default:
		return 0, false
	}
}
