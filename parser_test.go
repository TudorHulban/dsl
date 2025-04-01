package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	dsl := `
	dataset "sales_data" {
	  criteria "high_volume" {
		baseline threshold = 1000;
		increment extra = 50;
	
		monitor "order_count" {
		  level 1 when value > threshold;
		  level 2 when value >= threshold + extra; // simple expression
		}
	  }
	  criteria "returns" {
		 monitor "return_rate_pct" {
			level 1 when value > 5.5;
		 }
	  }
	}
	`

	reader := strings.NewReader(dsl)

	ast, errors := parse(reader)
	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "Parsing Errors:")
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, "-", err)
		}

		t.FailNow()
	}

	if ast != nil {
		fmt.Println("Parsing Successful (AST generated):")
		// You would now traverse the 'ast' (*program) object
		// to create your alerts.
		// Example: Print dataset names
		for _, ds := range ast.datasets {
			fmt.Printf("  Dataset: %s\n", ds.name)
			for _, crit := range ds.criteria {
				fmt.Printf("    Criteria: %s\n", crit.name)
				// ... traverse deeper ...
			}
		}
	} else if len(errors) == 0 {
		fmt.Fprintln(os.Stderr, "Parsing failed, AST is nil, but no specific errors reported (unexpected state).")
	}
}
