package s3bucket

import (
	"testing"
)

func Test_ContainsBucketState(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description      string
		installation     string
		bucketNameToFind string
		bucketStateList  []BucketState
		expectedValue    bool
	}{
		{
			description:      "basic match",
			installation:     "test-install",
			bucketNameToFind: "bck1",
			bucketStateList:  []BucketState{},
			expectedValue:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := containsBucketState(tc.bucketNameToFind, tc.bucketStateList)

			if result != tc.expectedValue {
				t.Fatalf("Expected %t tags, found %t", tc.expectedValue, result)
			}
		})
	}
}
