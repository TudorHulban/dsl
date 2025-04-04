package dslalert

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMonitorParsing(t *testing.T) {
	t.Run(
		"1. error - empty monitor block",
		func(t *testing.T) {
			input := `monitor "orders" {}`

			p := newParser(&paramsNewParser{
				Lexer:       newLexer(strings.NewReader(input)),
				IsDebugMode: true,
			})

			_ = p.parseMonitor()

			require.NotNil(t, p.errors, "should reject empty monitor")
			require.Regexp(t,
				regexp.MustCompile(`(?i)(must contain at least one level rule|empty monitor|expected level rule)`),
				p.errors[0],
				"should reject empty monitor blocks\nGot: %q",
				p.errors[0],
			)
		},
	)

	t.Run(
		"2. error - unclosed monitor block",
		func(t *testing.T) {
			input := `monitor "orders" { level 1 when value > 5;` // Missing }

			p := newParser(
				&paramsNewParser{
					Lexer:       newLexer(strings.NewReader(input)),
					IsDebugMode: true,
				},
			)

			_ = p.parseMonitor()

			require.NotEmpty(t, p.errors, "should report error")
			require.Regexp(t,
				regexp.MustCompile(
					`expected [}]|token 11|not properly closed|missing closing brace`,
				),
				strings.ToLower(p.errors[0]),
				"should detect unclosed block\nGot: %q",
				p.errors[0],
			)
		},
	)

	t.Run(
		"3. error - missing monitor name",
		func(t *testing.T) {
			input := `monitor { level 1 when value > 5; }`

			p := newParser(
				&paramsNewParser{
					Lexer:       newLexer(strings.NewReader(input)),
					IsDebugMode: true,
				},
			)

			_ = p.parseMonitor()

			require.NotEmpty(t, p.errors, "should report error")
			require.Regexp(t,
				regexp.MustCompile(`(?i)(expected\s+string|missing\s+monitor\s+name|token\s+8)`),
				p.errors[0],
				"should complain about missing name\nGot: %q",
				p.errors[0],
			)
		},
	)

	t.Run(
		"4. valid monitor with rule",
		func(t *testing.T) {
			input := `monitor "orders" {
				level 1 when value > 100;
			}`

			p := newParser(
				&paramsNewParser{
					Lexer:       newLexer(strings.NewReader(input)),
					IsDebugMode: true,
				},
			)

			monitor := p.parseMonitor()

			require.Empty(t, p.errors)
			require.Equal(t, "orders", monitor.ColumnName)
			require.Len(t, monitor.Rules, 1, "should parse the rule")
		},
	)

	t.Run(
		"5. valid monitor with rules",
		func(t *testing.T) {
			input := `
		monitor "orders" {
			level 1 when value > 5;
			level 2 when value > 10;
		}
		`

			p := newParser(
				&paramsNewParser{
					Lexer:       newLexer(strings.NewReader(input)),
					IsDebugMode: true,
				},
			)

			monitor := p.parseMonitor()

			require.Empty(t, p.errors)
			require.Equal(t, "orders", monitor.ColumnName)
			require.Len(t, monitor.Rules, 2, "should parse both rules")
			require.Equal(t, 1, monitor.Rules[0].Level)
			require.Contains(t, monitor.Rules[0].Condition.string(), "value > 5")
		},
	)
}
