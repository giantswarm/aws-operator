package cloudformation

import (
	"context"
	"testing"

	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_Cloudformation_newDelete(t *testing.T) {
	testCases := []struct {
		obj               interface{}
		currentState      interface{}
		desiredState      interface{}
		expectedStackName string
		description       string
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
			currentState:      StackState{},
			desiredState:      StackState{},
			expectedStackName: "",
		},
		{
			description: "current state empty, desired state not empty, expected empty",
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
			expectedStackName: "",
		},
		{
			description: "current state not empty, desired state not empty but different, expected current state",
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
			expectedStackName: "current",
		},
		{
			description: "current state not empty, desired state not empty but equal, expected desired state",
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
				Name: "current",
			},
			expectedStackName: "",
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
			result, err := newResource.newDeleteChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			deleteChange, ok := result.(awscloudformation.DeleteStackInput)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteChange, result)
			}
			if *deleteChange.StackName != tc.expectedStackName {
				t.Errorf("expected %s, got %s", tc.expectedStackName, deleteChange.StackName)
			}
		})
	}
}
