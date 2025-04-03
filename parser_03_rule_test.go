package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleParsing(t *testing.T) {
	t.Run(
		"1. error - missing semicolon",
		func(t *testing.T) {
			input := `level 1 when value > 5`

			p := NewParser(
				&ParamsNewParser{
					Lexer:       newLexer(strings.NewReader(input)),
					IsDebugMode: true,
				},
			)

			_ = p.parseRule()

			require.NotNil(t, p.errors, "should report missing semicolon")
			require.Contains(t, p.errors[0], "expected ;", "error should be helpful")
		},
	)

	// t.Run(
	// 	"2. error - invalid level",
	// 	func(t *testing.T) {
	// 		input := `level abc when value > 5;`

	// 		p := NewParser(
	// 			&ParamsNewParser{
	// 				Lexer:       newLexer(strings.NewReader(input)),
	// 				IsDebugMode: true,
	// 			},
	// 		)

	// 		_ = p.parseRule()

	// 		require.NotNil(t, p.errors, "should report invalid level")
	// 		require.Contains(t, p.errors[0], "invalid level number", "error should be helpful")
	// 	},
	// )

	// t.Run(
	// 	"3. valid rule with simple condition",
	// 	func(t *testing.T) {
	// 		input := `level 1 when value > 5;`

	// 		p := NewParser(
	// 			&ParamsNewParser{
	// 				Lexer:       newLexer(strings.NewReader(input)),
	// 				IsDebugMode: true,
	// 			},
	// 		)

	// 		rule := p.parseRule()

	// 		require.Nil(t, p.errors, "should have no errors")
	// 		require.Equal(t, 1, rule.Level, "level should be 1")
	// 		require.Contains(t, rule.Condition.String(), "value > 5", "condition mismatch")
	// 	},
	// )
}
