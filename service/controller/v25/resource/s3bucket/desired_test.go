package s3bucket

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
)

func Test_Resource_S3Bucket_GetDesiredState(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj           interface{}
		expectedNames []string
		description   string
	}{
		{
			description: "Get bucket name from custom object.",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			expectedNames: []string{
				"5xchu-g8s-access-logs",
				"myaccountid-g8s-5xchu",
			},
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
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := controllercontext.NewContext(context.Background(), testContextWithAccountID("myaccountid"))

			result, err := newResource.GetDesiredState(ctx, tc.obj)
			if err != nil {
				t.Fatalf("expected '%v' got '%#v'", nil, err)
			}

			desiredBuckets, ok := result.([]BucketState)
			if !ok {
				t.Fatalf("case expected '%T', got '%T'", desiredBuckets, result)
			}

			// Order should be respected in the slice returned (always delivery log bucket first)
			for key, desiredBucket := range desiredBuckets {
				if tc.expectedNames[key] != desiredBucket.Name {
					t.Fatalf("expected bucket name %q got %q", tc.expectedNames[key], desiredBucket.Name)
				}
			}
		})
	}
}

func testContextWithAccountID(id string) controllercontext.Context {
	return controllercontext.Context{
		Status: controllercontext.ContextStatus{
			TenantCluster: controllercontext.ContextStatusTenantCluster{
				AWSAccountID: id,
			},
		},
	}
}
