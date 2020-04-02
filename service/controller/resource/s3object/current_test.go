package s3object

import (
	"context"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func Test_CurrentState(t *testing.T) {
	t.Parallel()
	clusterTpo := &providerv1alpha1.AWSConfig{
		Spec: providerv1alpha1.AWSConfigSpec{
			Cluster: providerv1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
			},
		},
	}

	testCases := []struct {
		obj             *providerv1alpha1.AWSConfig
		description     string
		expectedS3Error bool
		expectedKey     string
		expectedBucket  string
		expectedBody    string
	}{
		{
			description:    "basic match",
			obj:            clusterTpo,
			expectedKey:    "cloudconfig/myversion/worker",
			expectedBucket: "myaccountid-g8s-test-cluster",
			expectedBody:   "mybody",
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

			cloudconfig := &CloudConfigMock{}
			release := &releasev1alpha1.Release{}
			clientset := fake.NewSimpleClientset(release)

			var err error
			var newResource *Resource
			{
				c := Config{
					CertsSearcher:      certstest.NewSearcher(certstest.Config{}),
					CloudConfig:        cloudconfig,
					G8sClient:          clientset,
					Logger:             microloggertest.New(),
					RandomKeysSearcher: randomkeystest.NewSearcher(),
					RegistryDomain:     "example.com",
				}

				newResource, err = New(c)
				if err != nil {
					t.Error("expected", nil, "got", err)
				}
			}

			cc := controllercontext.Context{
				Client: controllercontext.ContextClient{
					TenantCluster: controllercontext.ContextClientTenantCluster{
						AWS: awsClients,
					},
				},
				Status: controllercontext.ContextStatus{
					TenantCluster: controllercontext.ContextStatusTenantCluster{
						AWSAccountID: "myaccountid",
					},
				},
			}
			ctx := context.TODO()
			ctx = controllercontext.NewContext(ctx, cc)

			result, err := newResource.GetCurrentState(ctx, tc.obj)
			if err != nil && !tc.expectedS3Error {
				t.Errorf("unexpected error %v", err)
			}
			if err == nil && tc.expectedS3Error {
				t.Error("expected S3 error didn't happen")
			}

			if !tc.expectedS3Error {
				currentState, ok := result.(map[string]BucketObjectState)
				if !ok {
					t.Errorf("expected '%T', got '%T'", currentState, result)
				}

				bucketObject, ok := currentState[tc.expectedKey]
				if !ok {
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
