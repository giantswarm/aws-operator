package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestHasStackChanged(t *testing.T) {
	t.Parallel()
	tests := []struct {
		params         map[string]string
		updatedParams  map[string]string
		expectedResult bool
	}{
		// Case 1. Updated parameters are the same.
		{
			params: map[string]string{
				"AZ":            "eu-central-1a",
				asgMaxSizeParam: "2",
				asgMinSizeParam: "2",
				imageIDParam:    "ami-test",
			},
			updatedParams: map[string]string{
				asgMaxSizeParam: "2",
				asgMinSizeParam: "2",
				imageIDParam:    "ami-test",
			},
			expectedResult: false,
		},
		// Case 2. ASG size parameters are different.
		{
			params: map[string]string{
				"AZ":            "eu-central-1a",
				asgMaxSizeParam: "2",
				asgMinSizeParam: "2",
				imageIDParam:    "ami-test",
			},
			updatedParams: map[string]string{
				asgMaxSizeParam: "3",
				asgMinSizeParam: "3",
				imageIDParam:    "ami-test",
			},
			expectedResult: true,
		},
		// Case 3. Image ID parameter is different.
		{
			params: map[string]string{
				"AZ":            "eu-central-1a",
				asgMaxSizeParam: "2",
				asgMinSizeParam: "2",
				imageIDParam:    "ami-test",
			},
			updatedParams: map[string]string{
				asgMaxSizeParam: "2",
				asgMinSizeParam: "2",
				imageIDParam:    "ami-new",
			},
			expectedResult: true,
		},
	}

	for i, tc := range tests {
		params := []*cloudformation.Parameter{}

		for k, v := range tc.params {
			param := &cloudformation.Parameter{}
			param.SetParameterKey(k)
			param.SetParameterValue(v)

			params = append(params, param)
		}

		result := hasStackChanged(params, tc.updatedParams)
		if result != tc.expectedResult {
			t.Fatalf("case %d expected hasStackChanged was '%t' but was '%t'", i, tc.expectedResult, result)
		}
	}
}
