package cloudformation

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"

	"github.com/giantswarm/aws-operator/service/awsconfig/v7/adapter"
	cloudformationservice "github.com/giantswarm/aws-operator/service/awsconfig/v7/cloudformation"
)

func Test_Resource_Cloudformation_newUpdateChange_updatesAllowed(t *testing.T) {
	clusterTpo := &v1alpha1.AWSConfig{
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

		c.Clients = &adapter.Clients{
			EC2: &adapter.EC2ClientMock{},
			IAM: &adapter.IAMClientMock{},
			KMS: &adapter.KMSClientMock{},
		}
		c.HostClients = &adapter.Clients{
			IAM: &adapter.IAMClientMock{},
			EC2: &adapter.EC2ClientMock{},
		}
		c.Logger = microloggertest.New()
		c.Service = &cloudformationservice.CloudFormation{}

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := updateallowedcontext.NewContext(context.Background(), make(chan struct{}))
			updateallowedcontext.SetUpdateAllowed(ctx)

			result, err := newResource.newUpdateChange(ctx, clusterTpo, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			updateChange, ok := result.(awscloudformation.UpdateStackInput)
			if !ok {
				t.Errorf("expected '%T', got '%T'", updateChange, result)
			}
			if updateChange.StackName != nil && *updateChange.StackName != *tc.expectedChange.StackName {
				t.Errorf("expected %v, got %v", *tc.expectedChange.StackName, *updateChange.StackName)
			}
		})
	}
}

func Test_Resource_Cloudformation_newUpdateChange_updatesNotAllowed(t *testing.T) {
	clusterTpo := &v1alpha1.AWSConfig{
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

		c.Clients = &adapter.Clients{
			EC2: &adapter.EC2ClientMock{},
			IAM: &adapter.IAMClientMock{},
			KMS: &adapter.KMSClientMock{},
		}
		c.HostClients = &adapter.Clients{
			IAM: &adapter.IAMClientMock{},
			EC2: &adapter.EC2ClientMock{},
		}
		c.Logger = microloggertest.New()
		c.Service = &cloudformationservice.CloudFormation{}

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), clusterTpo, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			updateChange, ok := result.(awscloudformation.UpdateStackInput)
			if !ok {
				t.Errorf("expected '%T', got '%T'", updateChange, result)
			}
			if updateChange.StackName != nil && *updateChange.StackName != *tc.expectedChange.StackName {
				t.Errorf("expected %v, got %v", *tc.expectedChange.StackName, *updateChange.StackName)
			}
		})
	}
}
