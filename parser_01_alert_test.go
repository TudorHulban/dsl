package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlertParsing(t *testing.T) {
	t.Run(
		"1. valid rule",
		func(t *testing.T) {
			inputValid := `
			criteria "c1" {
				monitor "column_orders" {
					level 1 when value > 5;
					level 2 when value > 10;
				}
			}
			`
			reader := strings.NewReader(inputValid)

			ast, errs := Parse(reader)

			require.Empty(t, errs, "should have no parsing errors")
			require.NotEmpty(t,
				ast.Criterias,
			)
			require.Equal(t,
				"c1",
				ast.Criterias[0].Name,
			)
			require.Len(t,
				ast.Criterias[0].Monitors,
				1,
			)
			require.Len(t,
				ast.Criterias[0].Monitors[0].Rules,
				2,
			)
			rule1 := ast.Criterias[0].Monitors[0].Rules[0]
			require.Regexp(t,
				regexp.MustCompile(
					`(?i)(value|5)`,
				),
				rule1.Condition.String(),
			)
			rule2 := ast.Criterias[0].Monitors[0].Rules[1]
			require.Regexp(t,
				regexp.MustCompile(
					`(?i)(value|10)`,
				),
				rule2.Condition.String(),
			)
		},
	)
}
