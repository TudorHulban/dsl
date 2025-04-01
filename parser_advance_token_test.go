package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestEmptyInput(t *testing.T) {
	inputEmpty := ""
	reader := strings.NewReader(inputEmpty)

	ast, errors := parse(reader)
	if len(errors) == 0 {
		t.Error("Empty input should return errors")
	}
	if ast != nil {
		t.Error("Empty input should return nil AST")
	}

	// Verify specific error message
	expectedError := "input is empty" // Match your actual error message
	found := false
	for _, err := range errors {
		if strings.Contains(err, expectedError) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected error containing '%s', got %v", expectedError, errors)
	}
}

func TestTokenAdvance(t *testing.T) {
	inputDataset := `dataset "test" { }`
	reader := strings.NewReader(inputDataset)

	ast, errors := parse(reader)
	if len(errors) > 0 {
		t.Fatalf("Unexpected errors: %v", errors)
	}
	if len(ast.datasets) == 0 {
		t.Fatal("Expected 1 dataset, got 0")
	}

	fmt.Printf("Parsed dataset: %+v\n", ast.datasets[0])
}
