package cloudformationv1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/clustertpr/spec/kubernetes"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/resource/cloudformationv1/adapter"
)

func TestMainTemplateGetEmptyBody(t *testing.T) {
	customObject := awstpr.CustomObject{}

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = adapter.Clients{}
	resourceConfig.Logger = microloggertest.New()

	newResource, err := New(resourceConfig)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = newResource.getMainTemplateBody(customObject)
	if err == nil {
		t.Error("error didn't happen")
	}
}

func TestMainTemplateExistingFields(t *testing.T) {
	// customObject with example fields for both asg and launch config
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Version: "myversion",
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
				Kubernetes: spec.Kubernetes{
					IngressController: kubernetes.IngressController{
						Domain: "mysubdomain.mydomain.com",
					},
				},
			},
			AWS: awsspec.AWS{
				AZ: "myaz",
				Workers: []awsspecaws.Node{
					awsspecaws.Node{
						ImageID: "myimageid",
					},
				},
			},
		},
	}

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
	}
	resourceConfig.Logger = microloggertest.New()

	newResource, err := New(resourceConfig)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	body, err := newResource.getMainTemplateBody(customObject)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !strings.Contains(body, "Description: Main CloudFormation stack.") {
		t.Error("stack header not found")
	}

	if !strings.Contains(body, "  workerLaunchConfiguration:") {
		t.Error("launch configuration header not found")
	}

	if !strings.Contains(body, "  workerAutoScalingGroup:") {
		t.Error("asg header not found")
	}

	if !strings.Contains(body, "ImageId: myimageid") {
		t.Error("launch configuration element not found")
	}

	if !strings.Contains(body, "AvailabilityZones: [myaz]") {
		fmt.Println(body)
		t.Error("asg element not found")
	}

	if !strings.Contains(body, "Outputs:") {
		fmt.Println(body)
		t.Error("outputs header not found")
	}

	if !strings.Contains(body, workersOutputKey+":") {
		fmt.Println(body)
		t.Error("workers output element not found")
	}
	if !strings.Contains(body, imageIDOutputKey+":") {
		fmt.Println(body)
		t.Error("imageID output element not found")
	}
	if !strings.Contains(body, clusterVersionOutputKey+":") {
		fmt.Println(body)
		t.Error("clusterVersion output element not found")
	}
	if !strings.Contains(body, "Value: myversion") {
		fmt.Println(body)
		t.Error("output element not found")
	}
}
