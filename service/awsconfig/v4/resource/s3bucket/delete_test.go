package s3bucket

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

func Test_Resource_S3Bucket_newDelete(t *testing.T) {
	testCases := []struct {
		obj                interface{}
		currentState       interface{}
		desiredState       interface{}
		expectedBucketName string
		description        string
	}{
		{
			description: "current and desired state empty, expected empty",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState:       BucketState{},
			desiredState:       BucketState{},
			expectedBucketName: "",
		},
		{
			description: "current state empty, desired state not empty, expected empty",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: BucketState{},
			desiredState: BucketState{
				Name: "desired",
			},
			expectedBucketName: "",
		},
		{
			description: "current state not empty, desired state not empty but equal, expected desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: BucketState{
				Name: "current",
			},
			desiredState: BucketState{
				Name: "current",
			},
			expectedBucketName: "current",
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
			result, err := newResource.newDeleteChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			deleteChange, ok := result.(BucketState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteChange, result)
			}
			if deleteChange.Name != tc.expectedBucketName {
				t.Errorf("expected %s, got %s", tc.expectedBucketName, deleteChange.Name)
			}
		})
	}
}
