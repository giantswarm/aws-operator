package s3bucket

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

func Test_Resource_S3Bucket_GetDesiredState(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj          interface{}
		expectedName string
		description  string
	}{
		{
			description: "Get bucket name from custom object.",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			expectedName: "000000000000-g8s-5xchu",
		},
	}

	var err error
	var awsService *awsservice.Service
	{
		awsConfig := awsservice.DefaultConfig()
		awsConfig.Clients = awsservice.Clients{
			STS: &awsservice.STSClientMock{},
		}
		awsConfig.Logger = microloggertest.New()
		awsService, err = awsservice.New(awsConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.AwsService = awsService
		resourceConfig.Logger = microloggertest.New()
		resourceConfig.InstallationName = "test-install"

		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatalf("expected '%v' got '%#v'", nil, err)
			}

			desiredBucket, ok := result.(BucketState)
			if !ok {
				t.Fatalf("case expected '%T', got '%T'", desiredBucket, result)
			}

			if tc.expectedName != desiredBucket.Name {
				t.Fatalf("expected bucket name '%s' got '%s'", tc.expectedName, desiredBucket.Name)
			}
		})
	}
}
