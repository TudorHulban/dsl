package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func parseExpr(input string) Expression {
	p := NewParser(
		&ParamsNewParser{
			Lexer:       newLexer(strings.NewReader(input)),
			IsDebugMode: true,
		},
	)

	expr := p.parseExpression(0)

	// Verify we consumed full input
	if p.tokenCurrent.kind != tokenEOF && len(p.errors) == 0 {
		p.errorf("unexpected trailing tokens")
	}

	if len(p.errors) > 0 {
		panic(fmt.Sprintf("parse error: %v", p.errors))
	}
	return expr
}

func TestOperatorPrecedence(t *testing.T) {
	t.Run(
		"operator precedence",
		func(t *testing.T) {
			input := "1 + 2 * 3"
			expr := parseExpr(input) // Helper to parse single expression
			require.Equal(t, "(1 + (2 * 3))", expr.String())
		},
	)

	t.Run(
		"comparison with math",
		func(t *testing.T) {
			input := "value > threshold + 5"
			expr := parseExpr(input)
			require.Equal(t, "(value > (threshold + 5))", expr.String())
		},
	)
}

func TestExpressionParsing(t *testing.T) {
	t.Run(
		"binary expression",
		func(t *testing.T) {
			input := "value > 5"

			p := NewParser(
				&ParamsNewParser{
					Lexer:       newLexer(strings.NewReader(input)),
					IsDebugMode: true,
				},
			)

			expr := p.parseExpression(0)

			require.IsType(t, &ExpressionBinary{}, expr)
			binExpr := expr.(*ExpressionBinary)
			require.Equal(t,
				"value",
				binExpr.LefthandSide.String(),
			)
			require.Equal(t,
				">",
				binExpr.Operator,
			)
			require.Equal(t,
				"5",
				binExpr.RighthandSide.String(),
			)
		},
	)
}
