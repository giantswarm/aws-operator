package create

import (
	"testing"
)

func Test_GetMaxBatchSize(t *testing.T) {
	tests := []struct {
		asgSize        int
		expectedResult int
	}{
		{
			// Case 1. Batch size should be 1 when less than 3 nodes.
			asgSize:        1,
			expectedResult: 1,
		},
		{
			// Case 2. Batch size should be 1 when less than 3 nodes.
			asgSize:        2,
			expectedResult: 1,
		},
		{
			// Case 3. Batch size should be 2 when less than 5 nodes.
			asgSize:        4,
			expectedResult: 2,
		},
		{
			// Case 4. Batch size should be 30% when over 4 nodes..
			asgSize:        7,
			expectedResult: 2,
		},
		{
			// Case 5. Batch size should be 30% when over 4 nodes..
			asgSize:        12,
			expectedResult: 4,
		},
	}

	for i, tc := range tests {
		if tc.expectedResult != getMaxBatchSize(tc.asgSize) {
			t.Fatalf("case %d expected max batch size to be %d but was %d", i+1, tc.expectedResult, getMaxBatchSize(tc.asgSize))
		}
	}
}

func Test_GetMinInstancesInService(t *testing.T) {
	tests := []struct {
		asgSize        int
		expectedResult int
	}{
		{
			// Case 1. Min instances should be 1 when nodes are 1.
			asgSize:        1,
			expectedResult: 1,
		},
		{
			// Case 2. Min instances should be 1 when nodes are 2.
			asgSize:        2,
			expectedResult: 1,
		},
		{
			// Case 3. Min instances should be 70% when nodes are 3.
			asgSize:        3,
			expectedResult: 2,
		},
		{
			// Case 4. Min instances should be 70% when nodes are 7.
			asgSize:        7,
			expectedResult: 5,
		},
		{
			// Case 5. Min instances should be 70% when nodes are 12.
			asgSize:        12,
			expectedResult: 8,
		},
	}

	for i, tc := range tests {
		if tc.expectedResult != getMinInstancesInService(tc.asgSize) {
			t.Fatalf("case %d expected min instances to be %d but was %d", i+1, tc.expectedResult, getMinInstancesInService(tc.asgSize))
		}
	}
}
