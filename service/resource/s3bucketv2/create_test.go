package s3bucketv2

import (
	"context"
	"testing"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_S3Bucket_newCreate(t *testing.T) {
	testCases := []struct {
		obj                interface{}
		currentState       interface{}
		desiredState       interface{}
		expectedBucketName string
		description        string
	}{
		{
			description: "current and desired state empty, expected empty",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			currentState:       BucketState{},
			desiredState:       BucketState{},
			expectedBucketName: "",
		},
		{
			description: "current state empty, desired state not empty, expected desired state",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			currentState: BucketState{},
			desiredState: BucketState{
				Name: "desired",
			},
			expectedBucketName: "desired",
		},
		{
			description: "current state not empty, desired state not empty but different, expected desired state",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			currentState: BucketState{
				Name: "current",
			},
			desiredState: BucketState{
				Name: "desired",
			},
			expectedBucketName: "desired",
		},
	}

	var err error
	var awsService *awsservice.Service
	{
		awsConfig := awsservice.DefaultConfig()
		awsConfig.Clients = awsservice.Clients{
			IAM: &awsservice.IAMClientMock{},
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
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Error("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newCreateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			createChange, ok := result.(BucketState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}
			if createChange.Name != tc.expectedBucketName {
				t.Errorf("expected %s, got %s", tc.expectedBucketName, createChange.Name)
			}
		})
	}
}
