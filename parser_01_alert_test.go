package main

import (
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

			config, errs := Parse(reader)

			require.Empty(t, errs, "should have no parsing errors")
			require.Len(t, config.Criterias, 1, "should parse criteria")
			require.Len(t,
				config.Criterias[0].Monitors,
				1,
				"should parse monitor",
			)

			rule := config.Criterias[0].Monitors[0].Rules[0]
			require.Contains(t,
				rule.Condition.String(),
				_dslCriteria,
				"condition should reference threshold",
			)
		},
	)
}
