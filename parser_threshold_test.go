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
			// Regex matches either error format
		},
	)

	t.Run(
		"2. error - wrong value type",
		func(t *testing.T) {
			inputMissingValue := `threshold = a`
			reader := strings.NewReader(inputMissingValue)

			_, errs := parse(reader)

			require.NotEmpty(t, errs, "should report syntax error")
			require.Contains(t, errs[0], "expected value", "error message should help debug")
		},
	)

	t.Run(
		"3. valid threshold declaration",
		func(t *testing.T) {
			inputValid := `threshold = 100`
			reader := strings.NewReader(inputValid)

			config, errs := parse(reader)

			require.Empty(t, errs, "should have no parsing errors")
			require.Len(t, config.Criterias, 1, "should create criteria")

			// Verify threshold is stored as a Setting
			settings := config.Criterias[0].Settings
			require.Len(t, settings, 1, "should parse threshold setting")
			require.Equal(t, "threshold", settings[0].Name)
			require.Equal(t, 100, settings[0].Value) // Assuming Value is int
		},
	)
}
