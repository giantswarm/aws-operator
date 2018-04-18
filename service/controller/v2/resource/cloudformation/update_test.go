package cloudformation

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/cloudformation/adapter"
)

func Test_Resource_Cloudformation_newUpdateChange(t *testing.T) {
	t.Parallel()
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
			description:  "current and desired state empty, expected empty",
			currentState: StackState{},
			desiredState: StackState{},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName:  aws.String(""),
				Parameters: []*awscloudformation.Parameter{},
			},
		},
		{
			description:  "current state empty, desired state not empty, expected empty",
			currentState: StackState{},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName:  aws.String(""),
				Parameters: []*awscloudformation.Parameter{},
			},
		},
		{
			description: "current state not empty, equal desired state, expected empty",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName:  aws.String(""),
				Parameters: []*awscloudformation.Parameter{},
			},
		},

		{
			description: "current state not empty, desired state not empty but different master image, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master-new",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "current state not empty, desired state not empty but different master CloudConfig version, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion-new",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "current state not empty, desired state not empty but different number of workers, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master-new",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "5",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "current state not empty, desired state not empty but different worker image, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker-new",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "current state not empty, desired state not empty but different worker CloudConfig version, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerCloudConfigVersion: "myWorkerCCVersion-new",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "current state not empty, desired state not empty but different master instance type, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterInstanceType:       "m1-large",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerInstanceType:       "m1-large",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master",
				MasterInstanceType:       "m1-xlarge",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerInstanceType:       "m1-large",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
		{
			description: "current state not empty, desired state not empty but different worker instance type, expected desired state",
			currentState: StackState{
				Name:                     "current",
				MasterImageID:            "ami-678-master",
				MasterCloudConfigVersion: "myMasterCCVersion",
				MasterInstanceType:       "m1-large",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerInstanceType:       "m1-large",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			desiredState: StackState{
				Name:                     "desired",
				MasterImageID:            "ami-678-master",
				MasterInstanceType:       "m1-large",
				MasterCloudConfigVersion: "myMasterCCVersion",
				WorkerCount:              "4",
				WorkerImageID:            "ami-1234-worker",
				WorkerInstanceType:       "m1-xlarge",
				WorkerCloudConfigVersion: "myWorkerCCVersion",
			},
			expectedChange: awscloudformation.UpdateStackInput{
				StackName: aws.String("desired"),
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.Clients = &adapter.Clients{
			EC2: &adapter.EC2ClientMock{},
			IAM: &adapter.IAMClientMock{},
			KMS: &adapter.KMSClientMock{},
		}
		resourceConfig.HostClients = &adapter.Clients{
			IAM: &adapter.IAMClientMock{},
			EC2: &adapter.EC2ClientMock{},
		}
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Error("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), clusterTpo, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
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
