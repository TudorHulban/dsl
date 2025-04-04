package dslalert_test

import (
	"fmt"
	"strings"
	dslalert "test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHowToUse(t *testing.T) {
	inputCriteria := `
			criteria "c1" {
				monitor "col1" {
					level 1 when value > 5;
					level 2 when value > 10;
				}
			}
			criteria "c2" {
				monitor "col2" {
					level 1 when value > 1;
					level 2 when value > 3;
					level 3 when value > 4.5;
				}
			}
			`

	inputDataset := []string{
		"customer_id,col1,col2",
		"1001,3,0",
		"1002,5,1",
		"1003,6,0",
		"1004,11,0",
		"1005,0,1",
		"1006,0,2",
		"1007,1,5",
		"1008,20,5",
		"1009,20.5,5.1",
	}

	reader := strings.NewReader(inputCriteria)

	ast, errorParse := dslalert.Parse(reader)
	require.Empty(t,
		errorParse,
		"should have no parsing errors",
	)

	for _, criteria := range ast.Criterias {
		resultsAlert, errEvaluate := dslalert.EvaluateCriteria(
			criteria,
			inputDataset,
		)
		require.NoError(t, errEvaluate)
		require.NotEmpty(t, resultsAlert)

		fmt.Printf(
			"\n%s\n%s",
			resultsAlert.Message(),
			resultsAlert,
		)

		fmt.Println(
			resultsAlert.Rows(),
		)
	}
}
