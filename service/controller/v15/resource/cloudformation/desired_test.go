package cloudformation

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v15/adapter"
)

func Test_Resource_Cloudformation_GetDesiredState(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj          interface{}
		expectedName string
		description  string
	}{
		{
			description: "CloudFormation gets name from custom object",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
					Cluster: v1alpha1.Cluster{
						ID:      "5xchu",
						Version: "cloud-formation",
					},
				},
			},
			expectedName: "cluster-5xchu-guest-main",
		},
	}

	var err error
	var newResource *Resource
	{
		c := Config{}

		c.HostClients = &adapter.Clients{}
		c.Logger = microloggertest.New()

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()

			result, err := newResource.GetDesiredState(ctx, tc.obj)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			desiredStack, ok := result.(StackState)
			if !ok {
				t.Fatalf("case expected '%T', got '%T'", desiredStack, result)
			}

			if tc.expectedName != desiredStack.Name {
				t.Fatalf("expected cloudformation name '%s' got '%s'", tc.expectedName, desiredStack.Name)
			}
		})
	}
}
