package s3objectv2

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr/certificatetprtest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeytpr/randomkeytprtest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

func Test_DesiredState(t *testing.T) {
	clusterTpo := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
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
			expectedMasterKey: "cloudconfig/myversion/master",
			expectedWorkerKey: "cloudconfig/myversion/worker",
		},
	}
	var err error
	var newResource *Resource

	resourceConfig := DefaultConfig()
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.AwsService = awsservice.AwsServiceMock{
		AccountID: "myaccountid",
		KeyArn:    "mykeyarn",
	}
	resourceConfig.Clients = Clients{
		KMS: &KMSClientMock{},
	}
	resourceConfig.CertWatcher = certificatetprtest.NewService()
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

			desiredState, ok := result.(BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", desiredState, result)
			}

			if desiredState.MasterCloudConfig.Bucket != tc.expectedBucket {
				t.Errorf("expected key %q, got %q", tc.expectedBucket, desiredState.MasterCloudConfig.Bucket)
			}

			if desiredState.MasterCloudConfig.Key != tc.expectedMasterKey {
				t.Errorf("expected key %q, got %q", tc.expectedMasterKey, desiredState.MasterCloudConfig.Key)
			}

			if desiredState.WorkerCloudConfig.Body != tc.expectedBody {
				t.Errorf("expected key %q, got %q", tc.expectedBody, desiredState.MasterCloudConfig.Body)
			}

			if desiredState.WorkerCloudConfig.Bucket != tc.expectedBucket {
				t.Errorf("expected key %q, got %q", tc.expectedBucket, desiredState.WorkerCloudConfig.Bucket)
			}

			if desiredState.WorkerCloudConfig.Key != tc.expectedWorkerKey {
				t.Errorf("expected key %q, got %q", tc.expectedWorkerKey, desiredState.WorkerCloudConfig.Key)
			}

			if desiredState.WorkerCloudConfig.Body != tc.expectedBody {
				t.Errorf("expected key %q, got %q", tc.expectedBody, desiredState.WorkerCloudConfig.Body)
			}
		})
	}
}
