package s3bucket

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
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

func Test_getS3BucketTags(t *testing.T) {
	testCases := []struct {
		obj          v1alpha1.AWSConfig
		includeTags  bool
		expectedTags []*s3.Tag
		description  string
	}{
		{
			description:  "do not include tags",
			obj:          v1alpha1.AWSConfig{},
			includeTags:  false,
			expectedTags: []*s3.Tag{},
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
	c.InstallationName = "installation-name"

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c.IncludeTags = tc.includeTags

			r, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			actual := r.getS3BucketTags(tc.obj)
			if !reflect.DeepEqual(actual, tc.expectedTags) {
				t.Errorf("tags don't match, want %#v, got %#v", tc.expectedTags, actual)
			}
		})
	}
}
