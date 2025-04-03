package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyInput(t *testing.T) {
	input := ""
	reader := strings.NewReader(input)

	ast, errors := Parse(reader)

	require.Nil(t, ast, "AST should be nil for empty input")
	require.NotEmpty(t, errors, "should return error for empty input")
	require.Regexp(t,
		regexp.MustCompile(`(?i)^input is empty$`), // Exact match, case-insensitive
		errors[0],

		"Expected empty input error\nGot: %q",
		errors[0],
	)
}

func TestTokenAdvance(t *testing.T) {
	input := `criteria "high_volume" { }`
	reader := strings.NewReader(input)

	ast, errors := Parse(reader)

	require.NotEmpty(t, errors)
	require.Nil(t, ast)
}
