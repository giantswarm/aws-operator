package adapter

import (
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
)

func TestAdapterInstanceRegularFields(t *testing.T) {
	testCases := []struct {
		description                    string
		customObject                   awstpr.CustomObject
		errorMatcher                   func(error) bool
		expectedAZ                     string
		expectedIAMInstanceProfileName string
		expectedImageID                string
		expectedInstanceType           string
		expectedSecurityGroupID        string
		expectedSubnetID               string
	}{
		{
			description:  "empty custom object",
			customObject: awstpr.CustomObject{},
			errorMatcher: IsInvalidConfig,
		},
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: defaultCluster,
					AWS: awsspec.AWS{
						AZ: "eu-central-1a",
						Masters: []awsspecaws.Node{
							awsspecaws.Node{
								ImageID:      "ami-test",
								InstanceType: "m3.large",
							},
						},
					},
				},
			},
			errorMatcher:                   nil,
			expectedAZ:                     "eu-central-1a",
			expectedIAMInstanceProfileName: "test-cluster-master-EC2-K8S-Role",
			expectedImageID:                "ami-test",
			expectedInstanceType:           "m3.large",
		},
	}

	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getInstance(tc.customObject, clients)
			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if a.instanceAdapter.AZ != tc.expectedAZ {
				t.Errorf("unexpected AZ, got %q, want %q", a.instanceAdapter.AZ, tc.expectedAZ)
			}

			if a.instanceAdapter.IAMInstanceProfileName != tc.expectedIAMInstanceProfileName {
				t.Errorf("unexpected IAMInstanceProfileName, got %q, want %q", a.instanceAdapter.AZ, tc.expectedAZ)
			}

			if a.instanceAdapter.ImageID != tc.expectedImageID {
				t.Errorf("unexpected ImageID, got %q, want %q", a.instanceAdapter.ImageID, tc.expectedAZ)
			}

			if a.instanceAdapter.InstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected InstanceType, got %q, want %q", a.instanceAdapter.InstanceType, tc.expectedInstanceType)
			}
		})
	}
}
