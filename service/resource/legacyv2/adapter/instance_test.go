package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterInstanceRegularFields(t *testing.T) {
	testCases := []struct {
		description             string
		customObject            v1alpha1.AWSConfig
		errorMatcher            func(error) bool
		expectedAZ              string
		expectedImageID         string
		expectedInstanceType    string
		expectedSecurityGroupID string
	}{
		{
			description:  "empty custom object",
			customObject: v1alpha1.AWSConfig{},
			errorMatcher: IsInvalidConfig,
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "eu-central-1a",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "ami-test",
								InstanceType: "m3.large",
							},
						},
					},
				},
			},
			errorMatcher:         nil,
			expectedAZ:           "eu-central-1a",
			expectedImageID:      "ami-test",
			expectedInstanceType: "m3.large",
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

			if a.MasterAZ != tc.expectedAZ {
				t.Errorf("unexpected MasterAZ, got %q, want %q", a.instanceAdapter.MasterAZ, tc.expectedAZ)
			}

			if a.MasterImageID != tc.expectedImageID {
				t.Errorf("unexpected MasterImageID, got %q, want %q", a.instanceAdapter.MasterImageID, tc.expectedAZ)
			}

			if a.MasterInstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected MasterInstanceType, got %q, want %q", a.instanceAdapter.MasterInstanceType, tc.expectedInstanceType)
			}
		})
	}
}
