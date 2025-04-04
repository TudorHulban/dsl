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
				monitor "orders" {
					level 1 when value > 5;
					level 2 when value > 10;
				}
			}
			criteria "c2" {
				monitor "returns" {
					level 1 when value > 1;
					level 2 when value > 3;
				}
			}
			`

	inputDataset := []string{
		"customer_id,orders,returns",
		"1001,3,0",
		"1002,5,1",
		"1003,6,0",
		"1004,11,0",
		"1005,0,1",
		"1006,0,2",
		"1007,1,5",
		"1008,20,5",
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

		fmt.Println(
			len(resultsAlert),
			resultsAlert,
		)
	}
}
