package s3objectv1

import (
	"context"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_CurrentState(t *testing.T) {
	clusterTpo := &awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
				Version: "myversion",
			},
		},
	}

	testCases := []struct {
		obj              *awstpr.CustomObject
		description      string
		expectedIAMError bool
		expectedS3Error  bool
		expectedKey      string
		expectedBucket   string
		expectedBody     string
	}{
		{
			description:    "basic match",
			obj:            clusterTpo,
			expectedKey:    "cloudconfig/myversion/worker",
			expectedBucket: "myaccountid-g8s-test-cluster",
			expectedBody:   "mybody",
		},
		{
			description:      "IAM error",
			obj:              clusterTpo,
			expectedIAMError: true,
		},

		{
			description:     "S3 error",
			obj:             clusterTpo,
			expectedS3Error: true,
		},
	}
	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resourceConfig.AwsService = AwsServiceMock{
				accountID: "myaccountid",
				isError:   tc.expectedIAMError,
			}
			resourceConfig.Clients = Clients{
				S3: &S3ClientMock{
					isError: tc.expectedS3Error,
					body:    tc.expectedBody,
				},
			}
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
			if err != nil && !tc.expectedIAMError && !tc.expectedS3Error {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if err == nil && tc.expectedIAMError {
				t.Error("expected IAM error didn't happen")
			}

			if err == nil && tc.expectedS3Error {
				t.Error("expected S3 error didn't happen")
			}

			if currentState.WorkerCloudConfig.Key != tc.expectedKey {
				t.Errorf("expeccted key %q, got %q", tc.expectedKey, currentState.WorkerCloudConfig.Key)
			}

			if currentState.WorkerCloudConfig.Bucket != tc.expectedBucket {
				t.Errorf("expeccted key %q, got %q", tc.expectedBucket, currentState.WorkerCloudConfig.Bucket)
			}

			if currentState.WorkerCloudConfig.Body != tc.expectedBody {
				t.Errorf("expeccted key %q, got %q", tc.expectedBody, currentState.WorkerCloudConfig.Body)
			}
		})
	}
}
