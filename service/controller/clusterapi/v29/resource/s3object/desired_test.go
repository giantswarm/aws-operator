package s3object

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
)

func Test_DesiredState(t *testing.T) {
	masterKeyPattern := "cloudconfig/v[\\d_]+/master"
	workerKeyPattern := "cloudconfig/v[\\d_]+/worker"

	masterKeyRegexp := regexp.MustCompile(masterKeyPattern)
	workerKeyRegexp := regexp.MustCompile(workerKeyPattern)

	testCases := []struct {
		obj            interface{}
		description    string
		expectedBucket string
		expectedBody   string
	}{
		{
			description: "basic match",
			obj: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						label.Cluster: "5xchu",
					},
				},
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
			},
			expectedBody:   "mybody-",
			expectedBucket: "myaccountid-g8s-5xchu",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			awsClients := aws.Clients{
				KMS: &KMSClientMock{},
			}

			cloudconfig := &CloudConfigMock{
				template: tc.expectedBody,
			}

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
					t.Fatal("expected", nil, "got", err)
				}
			}

			ctx := controllercontext.NewContext(context.Background(), controllercontext.Context{
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
			})

			result, err := newResource.GetDesiredState(ctx, tc.obj)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			desiredState, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", desiredState, result)
			}

			if len(desiredState) != 2 {
				t.Fatalf("expected 2 objects, got %d", len(desiredState))
			}

			for key, bucketObjectState := range desiredState {
				if bucketObjectState.Bucket != tc.expectedBucket {
					t.Fatalf("expected bucket %q, got %q", tc.expectedBucket, bucketObjectState.Bucket)
				}

				if bucketObjectState.Body != tc.expectedBody {
					t.Fatalf("expected body %q, got %q", tc.expectedBody, bucketObjectState.Body)
				}

				if strings.HasSuffix(key, "master") {
					if !masterKeyRegexp.MatchString(key) {
						t.Fatalf("expected key %q, to match pattern %q", key, masterKeyPattern)
					}
				} else if strings.HasSuffix(key, "worker") {
					if !workerKeyRegexp.MatchString(key) {
						t.Fatalf("expected key %q, to match pattern %q", key, workerKeyPattern)
					}
				} else {
					t.Fatalf("unexpected key %q", key)
				}
			}
		})
	}
}
