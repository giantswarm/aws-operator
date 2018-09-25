package cloudformation

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v14patch3/adapter"
)

func Test_Resource_Cloudformation_newDelete(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj               interface{}
		currentState      interface{}
		desiredState      interface{}
		expectedStackName string
		description       string
	}{
		{
			description: "case 0: current and desired state empty, expected empty",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState:      StackState{},
			desiredState:      StackState{},
			expectedStackName: "",
		},
		{
			description: "case 1: current state empty, desired state not empty, expected desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: StackState{},
			desiredState: StackState{
				Name: "desired",
			},
			expectedStackName: "desired",
		},
		{
			description: "case 2: current state not empty, desired state not empty but different, expected desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: StackState{
				Name: "current",
			},
			desiredState: StackState{
				Name: "desired",
			},
			expectedStackName: "desired",
		},
		{
			description: "case 3: current state not empty, desired state not empty but equal, expected desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			currentState: StackState{
				Name: "equal",
			},
			desiredState: StackState{
				Name: "equal",
			},
			expectedStackName: "equal",
		},
	}

	var err error
	var newResource *Resource
	{
		c := Config{}

		c.HostClients = &adapter.Clients{}
		c.Logger = microloggertest.New()
		c.EncrypterBackend = "kms"

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()

			result, err := newResource.newDeleteChange(ctx, tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			deleteChange, ok := result.(StackState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", deleteChange, result)
			}
			if deleteChange.Name != tc.expectedStackName {
				t.Fatalf("expected %s, got %s", tc.expectedStackName, deleteChange.Name)
			}
		})
	}
}
