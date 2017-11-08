package cloudformation

import (
	"context"
	"testing"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_Cloudformation_newCreate(t *testing.T) {
	testCases := []struct {
		obj                  interface{}
		currentState         interface{}
		desiredState         interface{}
		expectedCreateChange StackState
		description          string
	}{
		{
			description: "current and desired state empty, expected empty",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			currentState:         StackState{},
			desiredState:         StackState{},
			expectedCreateChange: StackState{},
		},
		{
			description: "current state empty, desired state not empty, expected desired state",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			currentState: StackState{},
			desiredState: StackState{
				Name: "desired",
			},
			expectedCreateChange: StackState{
				Name: "desired",
			},
		},
		{
			description: "current state not empty, desired state not empty but different, expected desired state",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			currentState: StackState{
				Name: "current",
			},
			desiredState: StackState{
				Name: "desired",
			},
			expectedCreateChange: StackState{
				Name: "desired",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		awsCfg := awsutil.Config{}
		resourceConfig.Clients = awsutil.NewClients(awsCfg)
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Error("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newCreateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}

			createChange, ok := result.(StackState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", createChange, result)
			}

			if createChange.Name != tc.expectedCreateChange.Name {
				t.Errorf("expected %s, got %s", tc.expectedCreateChange.Name, createChange.Name)
			}
		})
	}
}
