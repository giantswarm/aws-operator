package cloudformation

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v14/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v14/controllercontext"
)

func Test_Resource_Cloudformation_newUpdateChange_updatesAllowed(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain: "mysubdomain.mydomain.com",
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain: "mysubdomain.mydomain.com",
					},
				},
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				AZ: "eu-central-1a",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					{},
				},
				Region: "eu-central-1",
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{},
				},
			},
		},
		Status: statusWithAllocatedSubnet("10.1.1.0/24"),
	}

	testCases := []struct {
		currentState   interface{}
		desiredState   interface{}
		expectedChange awscloudformation.UpdateStackInput
		description    string
	}{
		{
			description:  "case 0, current state empty, desired state empty, expected empty state",
			currentState: StackState{},
			desiredState: StackState{},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description:  "case 1, current state empty, desired state not empty, expected desired state",
			currentState: StackState{},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 2, current state not empty, equal desired state, expected empty",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 3, current state not empty, desired state not empty, different master image, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "CHANGED",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 4, current state not empty, desired state not empty, different master CloudConfig version, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "CHANGED",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 5, current state not empty, desired state not empty, different number of workers, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "CHANGED",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 6, current state not empty, desired state not empty, different worker image, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "CHANGED",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 7, current state not empty, desired state not empty, different worker CloudConfig version, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "CHANGED",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 8, current state not empty, desired state not empty, different master instance type, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "CHANGED",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 9, current state not empty, desired state not empty, different worker instance type, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "CHANGED",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 10, current state not empty, desired state not empty, different version bundle version, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "CHANGED",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
	}

	var err error
	var newResource *Resource
	{
		c := Config{}

		c.HostClients = &adapter.Clients{
			IAM: &adapter.IAMClientMock{},
			EC2: &adapter.EC2ClientMock{},
			STS: &adapter.STSClientMock{},
		}
		c.Logger = microloggertest.New()
		c.EncrypterBackend = "kms"
		c.GuestPrivateSubnetMaskBits = 25
		c.GuestPublicSubnetMaskBits = 25

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	awsClients := awsclient.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		STS: &adapter.STSClientMock{},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := updateallowedcontext.NewContext(context.Background(), make(chan struct{}))
			ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})
			updateallowedcontext.SetUpdateAllowed(ctx)

			result, err := newResource.newUpdateChange(ctx, customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			updateChange, ok := result.(StackState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", updateChange, result)
			}
			if updateChange.UpdateStackInput.StackName != nil && *updateChange.UpdateStackInput.StackName != *tc.expectedChange.StackName {
				t.Fatalf("expected %v, got %v", *tc.expectedChange.StackName, *updateChange.UpdateStackInput.StackName)
			}
		})
	}
}

func Test_Resource_Cloudformation_newUpdateChange_updatesNotAllowed(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain: "mysubdomain.mydomain.com",
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain: "mysubdomain.mydomain.com",
					},
				},
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				AZ: "eu-central-1a",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					{},
				},
				Region: "eu-central-1",
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{},
				},
			},
		},
		Status: statusWithAllocatedSubnet("10.1.1.0/24"),
	}

	testCases := []struct {
		currentState   interface{}
		desiredState   interface{}
		expectedChange awscloudformation.UpdateStackInput
		description    string
	}{
		{
			description:  "case 0, current state empty, desired state empty, expected empty state",
			currentState: StackState{},
			desiredState: StackState{},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description:  "case 1, current state empty, desired state not empty, expected empty state",
			currentState: StackState{},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 2, current state not empty, equal desired state, expected empty",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 3, current state not empty, desired state not empty, different master image, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "CHANGED",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 4, current state not empty, desired state not empty, different master CloudConfig version, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "CHANGED",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 5, current state not empty, desired state not empty, different number of workers, expected desired state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "CHANGED",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "case 6, current state not empty, desired state not empty, different worker image, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "CHANGED",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 7, current state not empty, desired state not empty, different worker CloudConfig version, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "CHANGED",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 8, current state not empty, desired state not empty, different master instance type, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "CHANGED",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 9, current state not empty, desired state not empty, different worker instance type, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "CHANGED",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 10, current state not empty, desired state not empty, different master version bundle version, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "CHANGED",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
		{
			description: "case 11, current state not empty, desired state not empty, different worker version bundle version, expected empty state",
			currentState: StackState{
				Name: "current",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "1.0.0",
			},
			desiredState: StackState{
				Name: "desired",

				MasterCloudConfigVersion: "1.0.0",
				MasterImageID:            "ami-123",
				MasterInstanceType:       "m3.large",

				WorkerCloudConfigVersion: "1.0.0",
				WorkerCount:              "4",
				WorkerImageID:            "ami-123",
				WorkerInstanceType:       "m3.large",

				VersionBundleVersion: "CHANGED",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String(""),
			},
		},
	}

	var err error
	var newResource *Resource
	{
		c := Config{}

		c.HostClients = &adapter.Clients{
			IAM: &adapter.IAMClientMock{},
			EC2: &adapter.EC2ClientMock{},
			STS: &adapter.STSClientMock{},
		}
		c.Logger = microloggertest.New()
		c.EncrypterBackend = "kms"
		c.GuestPrivateSubnetMaskBits = 25
		c.GuestPublicSubnetMaskBits = 25

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	awsClients := awsclient.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		STS: &adapter.STSClientMock{},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})

			result, err := newResource.newUpdateChange(ctx, customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			updateChange, ok := result.(StackState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", updateChange, result)
			}
			if updateChange.UpdateStackInput.StackName != nil && *updateChange.UpdateStackInput.StackName != *tc.expectedChange.StackName {
				t.Fatalf("expected %v, got %v", *tc.expectedChange.StackName, *updateChange.UpdateStackInput.StackName)
			}
		})
	}
}
