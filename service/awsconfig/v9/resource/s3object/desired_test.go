package s3object

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy/legacytest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeytpr/randomkeytprtest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

func Test_DesiredState(t *testing.T) {
	t.Parallel()
	clusterTpo := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		obj               *v1alpha1.AWSConfig
		description       string
		expectedBucket    string
		expectedBody      string
		expectedMasterKey string
		expectedWorkerKey string
	}{
		{
			description:       "basic match",
			obj:               clusterTpo,
			expectedBody:      "mybody-",
			expectedBucket:    "myaccountid-g8s-test-cluster",
			expectedMasterKey: "cloudconfig/v_3_2_3/master",
			expectedWorkerKey: "cloudconfig/v_3_2_3/worker",
		},
	}
	var err error
	var newResource *Resource
	var masterCloudConfig BucketObjectState
	var workerCloudConfig BucketObjectState

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.AwsService = awsservice.AwsServiceMock{
		AccountID: "myaccountid",
		KeyArn:    "mykeyarn",
	}
	resourceConfig.Clients = Clients{
		KMS: &KMSClientMock{},
	}
	resourceConfig.CertWatcher = legacytest.NewService()
	resourceConfig.RandomKeyWatcher = randomkeytprtest.NewService()

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resourceConfig.CloudConfig = &CloudConfigMock{
				template: tc.expectedBody,
			}
			newResource, err = New(resourceConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			desiredState, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", desiredState, result)
			}

			if len(desiredState) != 2 {
				t.Errorf("expected 2 objects, got %d", len(desiredState))
			}

			if masterCloudConfig, ok = desiredState[tc.expectedMasterKey]; !ok {
				t.Errorf("expected key %q, not found", tc.expectedMasterKey)
			}

			if masterCloudConfig.Bucket != tc.expectedBucket {
				t.Errorf("expected bucket %q, got %q", tc.expectedBucket, masterCloudConfig.Bucket)
			}

			if masterCloudConfig.Key != tc.expectedMasterKey {
				t.Errorf("expected key %q, got %q", tc.expectedMasterKey, masterCloudConfig.Key)
			}

			if masterCloudConfig.Body != tc.expectedBody {
				t.Errorf("expected key %q, got %q", tc.expectedBody, masterCloudConfig.Body)
			}

			if workerCloudConfig, ok = desiredState[tc.expectedWorkerKey]; !ok {
				t.Errorf("expected key %q, not found", tc.expectedWorkerKey)
			}

			if workerCloudConfig.Bucket != tc.expectedBucket {
				t.Errorf("expected bucket %q, got %q", tc.expectedBucket, workerCloudConfig.Bucket)
			}

			if workerCloudConfig.Key != tc.expectedWorkerKey {
				t.Errorf("expected key %q, got %q", tc.expectedWorkerKey, workerCloudConfig.Key)
			}

			if workerCloudConfig.Body != tc.expectedBody {
				t.Errorf("expected key %q, got %q", tc.expectedBody, workerCloudConfig.Body)
			}
		})
	}
}
