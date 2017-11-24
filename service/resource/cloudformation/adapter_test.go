package cloudformation

import (
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

func TestAdapterMain(t *testing.T) {
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			AWS: awsspec.AWS{
				Workers: []awsspecaws.Node{
					awsspecaws.Node{},
				},
			},
		},
	}
	clients := awsutil.Clients{}

	a := adapter{}
	err := a.getMain(customObject, clients)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	expected := "worker"
	actual := a.ASGType

	if expected != actual {
		t.Errorf("unexpected value, expecting %q, got %q", expected, actual)
	}
}

func TestAdapterLaunchConfigurationRegularFields(t *testing.T) {
	testCases := []struct {
		description                      string
		customObject                     awstpr.CustomObject
		expectedError                    bool
		expectedImageID                  string
		expectedInstanceType             string
		expectedIAMInstanceProfileName   string
		expectedAssociatePublicIPAddress bool
	}{
		{
			description:   "empty custom object",
			customObject:  awstpr.CustomObject{},
			expectedError: true,
		},
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Workers: []awsspecaws.Node{
							awsspecaws.Node{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedImageID:      "myimageid",
			expectedInstanceType: "myinstancetype",
		},
	}

	clients := awsutil.Clients{}
	a := adapter{}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := a.getLaunchConfiguration(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.ImageID != tc.expectedImageID {
				t.Errorf("unexpected output, got %q, want %q", a.ImageID, tc.expectedImageID)
			}
			if a.InstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected output, got %q, want %q", a.InstanceType, tc.expectedInstanceType)
			}
		})
	}
}

func TestAdapterAutoScalingGroupRegularFields(t *testing.T) {
	testCases := []struct {
		description   string
		customObject  awstpr.CustomObject
		expectedError bool
		expectedAZ    string
	}{
		{
			description:  "empty custom object",
			customObject: awstpr.CustomObject{},
			expectedAZ:   "",
		},
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Region: "myregion",
					},
				},
			},
			expectedAZ: "myregion",
		},
	}

	clients := awsutil.Clients{}

	a := adapter{}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.AZ != tc.expectedAZ {
				t.Errorf("unexpected output, got %q, want %q", a.AZ, tc.expectedAZ)
			}
		})
	}
}
