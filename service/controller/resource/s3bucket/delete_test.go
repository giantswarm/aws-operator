package s3bucket

import (
	"context"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func Test_Resource_S3Bucket_newDelete(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj                interface{}
		currentState       []BucketState
		desiredState       []BucketState
		expectedBucketName string
		description        string
	}{
		{
			description: "current and desired state empty, expected empty",
			obj: &infrastructurev1alpha2.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.Cluster: "5xchu",
					},
				},
			},
			currentState:       []BucketState{},
			desiredState:       []BucketState{},
			expectedBucketName: "",
		},
		{
			description: "current state empty, desired state not empty, expected empty",
			obj: &infrastructurev1alpha2.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.Cluster: "5xchu",
					},
				},
			},
			currentState: []BucketState{},
			desiredState: []BucketState{
				{
					Name: "desired",
				},
			},
			expectedBucketName: "",
		},
		{
			description: "current state not empty, desired state not empty but equal, expected desired state avoiding delivery log bucket",
			obj: &infrastructurev1alpha2.AWSCluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.Cluster: "5xchu",
					},
				},
			},
			currentState: []BucketState{
				{
					Name: "current",
				},
				{
					Name:            "log-bucket",
					IsLoggingBucket: true,
				},
			},
			desiredState: []BucketState{
				{
					Name: "current",
				},
				{
					Name:            "log-bucket",
					IsLoggingBucket: true,
				},
			},
			expectedBucketName: "current",
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
			result, err := newResource.newDeleteChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}

			deleteChanges, ok := result.([]BucketState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteChanges, result)
			}

			for _, deleteChange := range deleteChanges {
				if deleteChange.Name != tc.expectedBucketName {
					t.Errorf("expected %s, got %s", tc.expectedBucketName, deleteChange.Name)
				}
			}
		})
	}
}
