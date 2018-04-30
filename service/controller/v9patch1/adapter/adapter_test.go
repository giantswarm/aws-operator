package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v9patch1/key"
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
		description           string
		customObject          v1alpha1.AWSConfig
		errorMatcher          func(error) bool
		expectedASGType       string
		expectedClusterID     string
		expectedMasterImageID string
		expectedWorkerImageID string
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
			},
			errorMatcher:      nil,
			expectedASGType:   "worker",
			expectedClusterID: "test-cluster",
			expectedImageID:   "ami-604e118b",
		},
		{
			description: "different region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ:     "eu-west-1a",
						Region: "eu-west-1",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
			},
			errorMatcher:      nil,
			expectedASGType:   "worker",
			expectedClusterID: "test-cluster",
			expectedImageID:   "ami-34237c4d",
		},
		{
			description: "invalid region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ:     "invalid-1a",
						Region: "invalid-1",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
			},
			errorMatcher: key.IsInvalidConfig,
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
				},
				HostClients: Clients{
					EC2: &EC2ClientMock{},
					IAM: &IAMClientMock{},
				},
				InstallationName: "myinstallation",
				HostAccountID:    "myHostAccountID",
				StackState: StackState{
					MasterImageID: "master-image-id",
					WorkerImageID: "worker-image-id",
				},
			}
			a, err := NewGuest(config)
			if tc.errorMatcher != nil && err == nil {
				t.Fatal("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Fatal("expected", true, "got", false)
			}

			if tc.expectedASGType != a.ASGType {
				t.Fatalf("unexpected ASG type, expected %q, got %q", tc.expectedASGType, a.ASGType)
			}

			if tc.expectedClusterID != a.ClusterID {
				t.Fatalf("unexpected cluster ID, expected %q, got %q", tc.expectedClusterID, a.ClusterID)
			}

			if tc.expectedWorkerImageID != a.WorkerImageID {
				t.Fatalf("unexpected WorkerImageID, expected %q, got %q", tc.expectedWorkerImageID, a.WorkerImageID)
			}
		})
	}
}
