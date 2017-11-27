package cloudformation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/micrologger/microloggertest"
)

func TestMainTemplateGetEmptyBody(t *testing.T) {
	customObject := awstpr.CustomObject{}

	resourceConfig := DefaultConfig()
	resourceConfig.Clients = Clients{}
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
	resourceConfig.Clients = Clients{
		EC2: &eC2ClientMock{sgExists: true},
		IAM: &iAMClientMock{},
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
}
