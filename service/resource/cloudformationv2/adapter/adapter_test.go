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
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: defaultCluster,
			AWS: v1alpha1.AWSConfigSpecAWS{
				AZ: "eu-central-1a",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						ImageID: "ami-test-master",
					},
				},
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						ImageID: "ami-test-worker",
					},
				},
			},
		},
	}
	expectedASGType := prefixWorker
	expectedClusterID := "test-cluster"
	expectedMasterImageID := "ami-test-master"
	expectedWorkerImageID := "ami-test-worker"

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

	cfg := Config{
		CustomObject:     customObject,
		Clients:          clients,
		InstallationName: "myinstallation",
		HostAccountID:    "myHostAccountID",
		HostClients:      hostClients,
	}
	a, err := NewGuest(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if expectedASGType != a.ASGType {
		t.Errorf("unexpected value, expecting %q, got %q", expectedASGType, a.ASGType)
	}

	if expectedClusterID != a.ClusterID {
		t.Errorf("unexpected value, expecting %q, got %q", expectedClusterID, a.ClusterID)
	}

	if expectedMasterImageID != a.MasterImageID {
		t.Errorf("unexpected MasterImageID, got %q, want %q", a.MasterImageID, expectedMasterImageID)
	}

	if expectedWorkerImageID != a.WorkerImageID {
		t.Errorf("unexpected WorkerImageID, got %q, want %q", a.WorkerImageID, expectedWorkerImageID)
	}
}
