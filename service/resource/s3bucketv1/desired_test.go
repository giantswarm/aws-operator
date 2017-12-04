package s3bucketv1

import (
	"context"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

func Test_Resource_S3Bucket_GetDesiredState(t *testing.T) {
	testCases := []struct {
		obj          interface{}
		expectedName string
		description  string
	}{
		{
			description: "Get bucket name from custom object.",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			expectedName: "5xchu-g8s-000000000000",
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.AwsConfig = awsutil.Config{
			AccessKeyID:     "accessKeyID",
			AccessKeySecret: "accessKeySecret",
		}
		resourceConfig.Clients = Clients{
			IAM: &IAMClientMock{},
		}
		resourceConfig.Logger = microloggertest.New()
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
