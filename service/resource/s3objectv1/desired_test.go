package s3objectv1

import (
	"context"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

func Test_DesiredState(t *testing.T) {
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
		obj            *awstpr.CustomObject
		description    string
		expectedKey    string
		expectedBucket string
		expectedBody   string
	}{
		{
			description:    "basic match",
			obj:            clusterTpo,
			expectedKey:    "cloudconfig/myversion/worker",
			expectedBucket: "test-cluster-g8s-myaccountid",
			expectedBody:   "mybody-",
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
	defaultAssetsBundle := make(map[certificatetpr.AssetsBundleKey][]byte)
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resourceConfig.CertWatcher = &CertWatcherMock{
				certs: defaultAssetsBundle,
			}
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

			if desiredState.WorkerCloudConfig.Key != tc.expectedKey {
				t.Errorf("expeccted key %q, got %q", tc.expectedKey, desiredState.WorkerCloudConfig.Key)
			}

			if desiredState.WorkerCloudConfig.Bucket != tc.expectedBucket {
				t.Errorf("expeccted key %q, got %q", tc.expectedBucket, desiredState.WorkerCloudConfig.Bucket)
			}

			if desiredState.WorkerCloudConfig.Body != tc.expectedBody {
				t.Errorf("expeccted key %q, got %q", tc.expectedBody, desiredState.WorkerCloudConfig.Body)
			}
		})
	}
}
