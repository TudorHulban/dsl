package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenAdvance(t *testing.T) {
	input := `dataset "test" { }`
	reader := strings.NewReader(input)

	l := newlexer(reader)
	p := NewParser(l)
	p.EnableDebug()

	fmt.Println("=== PARSER DEBUG ===")
	p.logTokenState() // Initial state

	// Parse and check results
	ast, errors := parse(reader)
	if len(errors) > 0 {
		t.Fatalf("Unexpected errors: %v", errors)
	}
	require.NotNil(t, ast)
	require.Equal(t, 1, len(ast.datasets), "Should have 1 dataset")
	require.Equal(t, "test", ast.datasets[0].name, "Dataset name mismatch")

	fmt.Printf("\n=== FINAL AST ===\n%+v\n", ast.datasets)

	t.Logf("Parsed dataset: %+v", ast.datasets[0])
}
