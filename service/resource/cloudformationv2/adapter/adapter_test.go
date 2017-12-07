package adapter

import (
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/clustertpr/spec/kubernetes"
)

var (
	defaultCluster = clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
		Kubernetes: spec.Kubernetes{
			IngressController: kubernetes.IngressController{
				Domain: "mysubdomain.mydomain.com",
			},
		},
	}
)

func TestAdapterMain(t *testing.T) {
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: defaultCluster,
			AWS: awsspec.AWS{
				Workers: []awsspecaws.Node{
					awsspecaws.Node{},
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
