package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

var (
	defaultCluster = v1alpha1.Cluster{
		ID: "test-cluster",
		Kubernetes: v1alpha1.ClusterKubernetes{
			API: v1alpha1.ClusterKubernetesAPI{
				Domain: "api.domain",
			},
			IngressController: v1alpha1.ClusterKubernetesIngressController{
				Domain: "ingress.domain",
			},
		},
		Etcd: v1alpha1.ClusterEtcd{
			Domain: "etcd.domain",
		},
	}
)

func TestAdapterGuestMain(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description              string
		customObject             v1alpha1.AWSConfig
		errorMatcher             func(error) bool
		expectedASGType          string
		expectedEC2ServiceDomain string
		expectedMasterImageID    string
		expectedWorkerImageID    string
	}{
		{
			description: "basic match",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ:     "eu-central-1a",
						Region: "eu-central-1",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
				Status: v1alpha1.AWSConfigStatus{
					Cluster: v1alpha1.StatusCluster{
						Network: v1alpha1.StatusClusterNetwork{
							CIDR: "10.1.4.0/24",
						},
					},
				},
			},
			errorMatcher:             nil,
			expectedASGType:          "worker",
			expectedEC2ServiceDomain: "ec2.amazonaws.com",
			expectedMasterImageID:    "master-image-id",
			expectedWorkerImageID:    "worker-image-id",
		},
		{
			description: "china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ:     "cn-north-1a",
						Region: "cn-north-1",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
				Status: v1alpha1.AWSConfigStatus{
					Cluster: v1alpha1.StatusCluster{
						Network: v1alpha1.StatusClusterNetwork{
							CIDR: "10.1.4.0/24",
						},
					},
				},
			},
			errorMatcher:             nil,
			expectedASGType:          "worker",
			expectedEC2ServiceDomain: "ec2.amazonaws.com.cn",
			expectedMasterImageID:    "master-image-id",
			expectedWorkerImageID:    "worker-image-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config := Config{
				CustomObject: tc.customObject,
				Clients: Clients{
					EC2: &EC2ClientMock{},
					IAM: &IAMClientMock{},
					KMS: &KMSClientMock{},
					ELB: &ELBClientMock{},
					STS: &STSClientMock{},
				},
				HostClients: Clients{
					EC2: &EC2ClientMock{},
					IAM: &IAMClientMock{},
					STS: &STSClientMock{},
				},
				InstallationName:      "myinstallation",
				HostAccountID:         "myHostAccountID",
				PrivateSubnetMaskBits: 25,
				PublicSubnetMaskBits:  25,
				StackState: StackState{
					MasterImageID: "master-image-id",
					WorkerImageID: "worker-image-id",
				},
			}
			a, err := NewGuest(config)
			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if tc.expectedASGType != a.Guest.AutoScalingGroup.ASGType {
				t.Fatalf("unexpected ASG type, expected %q, got %q", tc.expectedASGType, a.Guest.AutoScalingGroup.ASGType)
			}
			if tc.expectedASGType != a.Guest.LaunchConfiguration.ASGType {
				t.Fatalf("unexpected ASG type, expected %q, got %q", tc.expectedASGType, a.Guest.LaunchConfiguration.ASGType)
			}

			if tc.expectedEC2ServiceDomain != a.Guest.IAMPolicies.EC2ServiceDomain {
				t.Fatalf("unexpected EC2 service domain, expected %q, got %q", tc.expectedEC2ServiceDomain, a.Guest.IAMPolicies.EC2ServiceDomain)
			}

			if tc.expectedWorkerImageID != a.Guest.LaunchConfiguration.WorkerImageID {
				t.Fatalf("unexpected WorkerImageID, expected %q, got %q", tc.expectedWorkerImageID, a.Guest.LaunchConfiguration.WorkerImageID)
			}
		})
	}
}
