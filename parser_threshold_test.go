package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThreshold(t *testing.T) {
	t.Run(
		"1. error - missing value",
		func(t *testing.T) {
			inputMissingValue := `threshold =`
			reader := strings.NewReader(inputMissingValue)

			_, errs := parse(reader)

			require.NotEmpty(t, errs, "should report syntax error")
			require.Regexp(t,
				`expected value|unexpected token`,
				errs[0],
				"error message should help debug",
			)
		},
	)

	t.Run(
		"2. error - wrong value type",
		func(t *testing.T) {
			inputMissingValue := `threshold = a`
			reader := strings.NewReader(inputMissingValue)

			_, errs := parse(reader)

			require.NotEmpty(t, errs, "should report syntax error")
			require.Regexp(t,
				`expected value|unexpected token|number`,
				errs[0],
				"error should mention value type",
			)
		},
	)

	t.Run(
		"3. valid rule",
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

			config, errs := parse(reader)

			require.Empty(t, errs, "should have no parsing errors")
			require.Len(t, config.Criterias, 1, "should parse criteria")
			require.Len(t, config.Criterias[0].Monitors, 1, "should parse monitor")

			rule := config.Criterias[0].Monitors[0].Rules[0]
			require.Contains(t,
				rule.Condition.String(),
				_dslCriteria,
				"condition should reference threshold",
			)
		},
	)
}
