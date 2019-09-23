package s3object

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v29patch1/controllercontext"
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

	masterKeyPattern := "cloudconfig/v[\\d_]+/master"
	workerKeyPattern := "cloudconfig/v[\\d_]+/worker"

	masterKeyRegexp := regexp.MustCompile(masterKeyPattern)
	workerKeyRegexp := regexp.MustCompile(workerKeyPattern)

	testCases := []struct {
		obj            *v1alpha1.AWSConfig
		description    string
		expectedBucket string
		expectedBody   string
	}{
		{
			description:    "basic match",
			obj:            clusterTpo,
			expectedBody:   "mybody-",
			expectedBucket: "myaccountid-g8s-test-cluster",
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

			c := controllercontext.Context{
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
			ctx = controllercontext.NewContext(ctx, c)

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
