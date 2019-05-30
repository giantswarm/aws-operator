package s3object

import (
	"context"
	"testing"

	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/controllercontext"
)

func Test_CurrentState(t *testing.T) {
	testCases := []struct {
		obj             interface{}
		description     string
		expectedS3Error bool
		expectedKey     string
		expectedBucket  string
		expectedBody    string
	}{
		{
			description: "basic match",
			obj: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProviderSpec: v1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: []byte(`
								{
									"cluster": {
										"versionBundle": {
											"version": "1.0.0"
										}
									}
								}
							`),
						},
					},
				},
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "test-cluster"
								}
							}
						`),
					},
				},
			},
			expectedKey:    "cloudconfig/myversion/worker",
			expectedBucket: "myaccountid-g8s-test-cluster",
			expectedBody:   "mybody",
		},
		{
			description: "S3 error",
			obj: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProviderSpec: v1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: []byte(`
								{
									"cluster": {
										"versionBundle": {
											"version": "1.0.0"
										}
									}
								}
							`),
						},
					},
				},
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "test-cluster"
								}
							}
						`),
					},
				},
			},
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

			var err error
			var newResource *Resource
			{
				c := Config{
					CertsSearcher:      certstest.NewSearcher(certstest.Config{}),
					CloudConfig:        cloudconfig,
					Logger:             microloggertest.New(),
					RandomKeysSearcher: randomkeystest.NewSearcher(),
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
