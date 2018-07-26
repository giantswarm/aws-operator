package s3object

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy/legacytest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"

	"github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v14/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v14/encrypter"
)

func Test_CurrentState(t *testing.T) {
	t.Parallel()
	clusterTpo := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
			},
		},
	}

	testCases := []struct {
		obj              *v1alpha1.AWSConfig
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

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			awsClients := aws.Clients{
				S3: &S3ClientMock{
					isError: tc.expectedS3Error,
					body:    tc.expectedBody,
				},
			}

			awsService := awsservice.AwsServiceMock{
				AccountID: "myaccountid",
				IsError:   tc.expectedIAMError,
			}

			cloudconfig := &CloudConfigMock{}

			var err error
			var newResource *Resource
			{
				c := Config{}
				c.CertWatcher = legacytest.NewService()
				c.Encrypter = &encrypter.EncrypterMock{}
				c.Logger = microloggertest.New()
				c.RandomKeySearcher = randomkeystest.NewSearcher()
				newResource, err = New(c)
				if err != nil {
					t.Error("expected", nil, "got", err)
				}
			}

			c := controllercontext.Context{
				AWSClient:   awsClients,
				AWSService:  awsService,
				CloudConfig: cloudconfig,
			}
			ctx := context.TODO()
			ctx = controllercontext.NewContext(ctx, c)

			result, err := newResource.GetCurrentState(ctx, tc.obj)
			if err != nil && !tc.expectedIAMError && !tc.expectedS3Error {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if err == nil && tc.expectedIAMError {
				t.Error("expected IAM error didn't happen")
			}

			if err == nil && tc.expectedS3Error {
				t.Error("expected S3 error didn't happen")
			}

			if !tc.expectedIAMError && !tc.expectedS3Error {
				var bucketObject BucketObjectState

				if bucketObject, ok = currentState[tc.expectedKey]; !ok {
					t.Errorf("expected S3 key %q not found", tc.expectedKey)
				}

				if bucketObject.Body != tc.expectedBody {
					t.Errorf("expected body %q, got %q", tc.expectedBody, bucketObject.Body)
				}

				if bucketObject.Bucket != tc.expectedBucket {
					t.Errorf("expected bucket %q, got %q", tc.expectedBucket, bucketObject.Bucket)
				}

				if bucketObject.Key != tc.expectedKey {
					t.Errorf("expected key %q, got %q", tc.expectedKey, bucketObject.Key)
				}
			}
		})
	}
}
