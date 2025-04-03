package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
