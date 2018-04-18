package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/awsconfig/v3/key"
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
		description       string
		customObject      v1alpha1.AWSConfig
		errorMatcher      func(error) bool
		expectedASGType   string
		expectedClusterID string
		expectedImageID   string
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
			expectedImageID:   "ami-90c152ff",
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
			expectedImageID:   "ami-32d1474b",
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

	clients := Clients{
		EC2: &EC2ClientMock{},
		IAM: &IAMClientMock{},
		KMS: &KMSClientMock{},
		ELB: &ELBClientMock{},
	}
	hostClients := Clients{
		EC2: &EC2ClientMock{},
		IAM: &IAMClientMock{},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject:     tc.customObject,
				Clients:          clients,
				InstallationName: "myinstallation",
				HostAccountID:    "myHostAccountID",
				HostClients:      hostClients,
			}
			a, err := NewGuest(cfg)
			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if tc.expectedASGType != a.ASGType {
				t.Errorf("unexpected value, expecting %q, got %q", tc.expectedASGType, a.ASGType)
			}

			if tc.expectedClusterID != a.ClusterID {
				t.Errorf("unexpected value, expecting %q, got %q", tc.expectedClusterID, a.ClusterID)
			}

			if tc.expectedImageID != a.MasterImageID {
				t.Errorf("unexpected MasterImageID, expecting %q, want %q", tc.expectedImageID, a.MasterImageID)
			}

			if tc.expectedImageID != a.WorkerImageID {
				t.Errorf("unexpected WorkerImageID, expecting %q, want %q", tc.expectedImageID, a.WorkerImageID)
			}
		})
	}
}
