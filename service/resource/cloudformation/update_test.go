package cloudformation

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

func Test_Resource_Cloudformation_newUpdateChange(t *testing.T) {
	testCases := []struct {
		obj            interface{}
		currentState   interface{}
		desiredState   interface{}
		expectedChange interface{}
		description    string
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
			currentState: StackState{},
			desiredState: StackState{},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
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
				Name:           "desired",
				Workers:        "4",
				ImageID:        "ami-1234",
				ClusterVersion: "myclusterversion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName:    aws.String("desired"),
				TemplateBody: aws.String(""),
				Parameters: []*awscloudformation.Parameter{
					{
						ParameterKey:   aws.String("workers"),
						ParameterValue: aws.String("4"),
					},
					{
						ParameterKey:   aws.String("imageID"),
						ParameterValue: aws.String("ami-1234"),
					},
					{
						ParameterKey:   aws.String("clusterVersion"),
						ParameterValue: aws.String("myclusterversion"),
					},
				},
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
				Name:           "current",
				Workers:        "3",
				ImageID:        "ami-6789",
				ClusterVersion: "oldclusterversion",
			},
			desiredState: StackState{
				Name:           "desired",
				Workers:        "4",
				ImageID:        "ami-1234",
				ClusterVersion: "myclusterversion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName:    aws.String("desired"),
				TemplateBody: aws.String(""),
				Parameters: []*awscloudformation.Parameter{
					{
						ParameterKey:   aws.String("workers"),
						ParameterValue: aws.String("4"),
					},
					{
						ParameterKey:   aws.String("imageID"),
						ParameterValue: aws.String("ami-1234"),
					},
					{
						ParameterKey:   aws.String("clusterVersion"),
						ParameterValue: aws.String("myclusterversion"),
					},
				},
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
			result, err := newResource.newUpdateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			updateChange, ok := result.(awscloudformation.UpdateStackInput)
			if !ok {
				t.Errorf("expected '%T', got '%T'", updateChange, result)
			}
			if !reflect.DeepEqual(updateChange, tc.expectedChange) {
				t.Errorf("expected %v, got %v", tc.expectedChange, updateChange)
			}
		})
	}
}
