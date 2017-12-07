package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

var (
	defaultCluster = v1alpha1.Cluster{
		ID: "test-cluster",
		Kubernetes: v1alpha1.ClusterKubernetes{
			IngressController: v1alpha1.ClusterKubernetesIngressController{
				Domain: "mysubdomain.mydomain.com",
			},
		},
	}
)

func TestAdapterMain(t *testing.T) {
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: defaultCluster,
			AWS: v1alpha1.AWSConfigSpecAWS{
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{},
				},
			},
		},
	}
	clients := Clients{
		EC2: &EC2ClientMock{},
		IAM: &IAMClientMock{},
	}

	a, err := New(customObject, clients)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	expected := prefixWorker
	actual := a.ASGType

	if expected != actual {
		t.Errorf("unexpected value, expecting %q, got %q", expected, actual)
	}
}
