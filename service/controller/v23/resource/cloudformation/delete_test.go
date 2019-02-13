package cloudformation

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v23/adapter"
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
			description: "case 1: current state not empty, desired state empty but different, expected current state",
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
			desiredState:      StackState{},
			expectedStackName: "current",
		},
		{
			description: "case 2: current state not empty, desired state not empty but different, expected current state",
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
			expectedStackName: "current",
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

		c.EncrypterBackend = "kms"
		c.G8sClient = fake.NewSimpleClientset()
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
