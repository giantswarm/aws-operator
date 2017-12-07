package cloudformationv2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2/adapter"
)

func TestMainTemplateGetEmptyBody(t *testing.T) {
	customObject := v1alpha1.AWSConfig{}

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
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
				Kubernetes: v1alpha1.ClusterKubernetes{
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain: "mysubdomain.mydomain.com",
					},
				},
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				AZ: "myaz",
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
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
