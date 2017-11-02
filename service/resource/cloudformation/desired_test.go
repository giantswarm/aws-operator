package cloudformation

import (
	"context"
	"testing"

	awsCF "github.com/aws/aws-sdk-go/service/cloudformation"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_Resource_Cloudformation_GetDesiredState(t *testing.T) {
	testCases := []struct {
		obj          interface{}
		expectedName string
		description  string
	}{
		{
			description: "CloudFormation gets name from custom object",
			obj: &awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "5xchu",
						},
					},
				},
			},
			expectedName: "5xchu-main",
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
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatalf("expected '%v' got '%#v'", nil, err)
			}

			desiredStack, ok := result.(*awsCF.CreateStackInput)
			if !ok {
				t.Fatalf("case expected '%T', got '%T'", desiredStack, result)
			}

			if tc.expectedName != *desiredStack.StackName {
				t.Fatalf("expected cloudformation name '%s' got '%s'", tc.expectedName, *desiredStack.StackName)
			}
		})
	}
}
