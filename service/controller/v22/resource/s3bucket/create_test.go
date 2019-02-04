package s3bucket

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_S3Bucket_newCreate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj                  interface{}
		currentState         interface{}
		desiredState         interface{}
		expectedBucketsState []BucketState
		description          string
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
			currentState:         []BucketState{},
			desiredState:         []BucketState{},
			expectedBucketsState: []BucketState{},
		},
		{
			description: "current state empty, desired state not empty, expected desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: []BucketState{},
			desiredState: []BucketState{
				{
					Name: "desired",
				},
			},
			expectedBucketsState: []BucketState{
				{
					Name: "desired",
				},
			},
		},
		{
			description: "current state not empty, desired state not empty but different, expected desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: []BucketState{
				{
					Name: "current",
				},
			},
			desiredState: []BucketState{
				{
					Name: "desired",
				},
			},
			expectedBucketsState: []BucketState{
				{
					Name: "desired",
				},
			},
		},
	}

	var err error

	var newResource *Resource
	{
		c := Config{
			Logger:           microloggertest.New(),
			InstallationName: "test-install",
		}

		newResource, err = New(c)
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
			createChanges, ok := result.([]BucketState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChanges, result)
			}
			for _, expectedBucketState := range tc.expectedBucketsState {
				if !containsBucketState(expectedBucketState.Name, createChanges) {
					t.Errorf("expected %v, got %v", expectedBucketState, createChanges)
				}
			}
		})
	}
}
