package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlertParsing(t *testing.T) {
	t.Run(
		"1. one criteria",
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

	t.Run(
		"2. two criterias",
		func(t *testing.T) {
			inputValid := `
			criteria "c1" {
				monitor "column_orders" {
					level 1 when value > 5;
					level 2 when value > 10;
				}
			}

			criteria "c2" {
				monitor "column_returns" {
					level 1 when value > 7;
					level 2 when value > 9;
				}
			}
			`
			reader := strings.NewReader(inputValid)

			ast, errs := Parse(reader)

			require.Empty(t, errs, "should have no parsing errors")
			require.Len(t,
				ast.Criterias,
				2,
			)
			require.Equal(t,
				"c1",
				ast.Criterias[0].Name,
			)
			require.Equal(t,
				"c2",
				ast.Criterias[1].Name,
			)
			require.Len(t,
				ast.Criterias[0].Monitors,
				1,
			)
			require.Len(t,
				ast.Criterias[1].Monitors,
				1,
			)
			require.Len(t,
				ast.Criterias[0].Monitors[0].Rules,
				2,
			)
			require.Len(t,
				ast.Criterias[1].Monitors[0].Rules,
				2,
			)

			criteria1Rule1 := ast.Criterias[0].Monitors[0].Rules[0]
			require.Regexp(t,
				regexp.MustCompile(
					`(?i).*value.*5.*`,
				),
				criteria1Rule1.Condition.String(),
			)
			criteria1Rule2 := ast.Criterias[0].Monitors[0].Rules[1]
			require.Regexp(t,
				regexp.MustCompile(
					`(?i).*value.*10.*`,
				),
				criteria1Rule2.Condition.String(),
			)

			criteria2Rule1 := ast.Criterias[1].Monitors[0].Rules[0]
			require.Regexp(t,
				regexp.MustCompile(
					`(?i).*value.*7.*`,
				),
				criteria2Rule1.Condition.String(),
			)
			criteria2Rule2 := ast.Criterias[1].Monitors[0].Rules[1]
			require.Regexp(t,
				regexp.MustCompile(
					`(?i).*value.*9.*`,
				),
				criteria2Rule2.Condition.String(),
			)
		},
	)
}
