package s3bucket

import (
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
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
func Test_BucketCanBeDeleted(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description         string
		installation        string
		deleteLoggingBucket bool
		bucketState         BucketState
		expectedValue       bool
	}{
		{
			description:         "test env true",
			installation:        "test-install",
			deleteLoggingBucket: true,
			bucketState:         BucketState{},
			expectedValue:       true,
		},
		{
			description:         "test env false",
			installation:        "test-install",
			deleteLoggingBucket: false,
			bucketState:         BucketState{},
			expectedValue:       true,
		},
		{
			description:         "test env true no logging bucket",
			installation:        "test-install",
			deleteLoggingBucket: true,
			bucketState:         BucketState{},
			expectedValue:       true,
		},
		{
			description:         "test env true logging bucket",
			installation:        "test-install",
			deleteLoggingBucket: true,
			bucketState: BucketState{
				IsLoggingBucket: true,
			},
			expectedValue: true,
		},
		{
			description:         "test env false no logging bucket",
			installation:        "test-install",
			deleteLoggingBucket: true,
			bucketState:         BucketState{},
			expectedValue:       true,
		},
		{
			description:         "test env false logging bucket",
			installation:        "test-install",
			deleteLoggingBucket: false,
			bucketState: BucketState{
				IsLoggingBucket: true,
			},
			expectedValue: false,
		},
	}

	var awsService *awsservice.Service
	{
		var err error
		awsConfig := awsservice.DefaultConfig()
		awsConfig.Clients = awsservice.Clients{}
		awsConfig.Logger = microloggertest.New()
		awsService, err = awsservice.New(awsConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	c := Config{}

	c.AwsService = awsService
	c.Logger = microloggertest.New()

	c.AccessLogsExpiration = 0

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c.DeleteLoggingBucket = tc.deleteLoggingBucket
			c.InstallationName = tc.installation
			r, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			result := r.canBeDeleted(tc.bucketState)

			if result != tc.expectedValue {
				t.Fatalf("Expected %t, found %t", tc.expectedValue, result)
			}
		})
	}
}
