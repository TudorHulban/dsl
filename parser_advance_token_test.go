package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyInput(t *testing.T) {
	inputEmpty := ""
	reader := strings.NewReader(inputEmpty)

	ast, errors := Parse(reader)
	require.Empty(t, errors)
	require.Nil(t, ast)

	// Verify specific error message
	expectedError := "input is empty" // Match your actual error message
	var found bool

	for _, err := range errors {
		if strings.Contains(err, expectedError) {
			found = true

			break
		}
	}

	if !found {
		t.Errorf(
			"Expected error containing '%s', got %v",
			expectedError,
			errors,
		)
	}
}

func TestTokenAdvance(t *testing.T) {
	inputDataset := `criteria "high_volume" { }`
	reader := strings.NewReader(inputDataset)

	ast, errors := Parse(reader)
	require.Empty(t, errors)
	require.NotEmpty(t,
		ast.Criterias,
	)

	fmt.Printf(
		"Parsed criteria: %+v\n",
		ast.Criterias[0],
	)
}
