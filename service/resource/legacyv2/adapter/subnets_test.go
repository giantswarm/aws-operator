package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterSubnetsRegularFields(t *testing.T) {
	testCases := []struct {
		description                              string
		customObject                             v1alpha1.AWSConfig
		expectedError                            bool
		expectedPrivateSubnetAZ                  string
		expectedPrivateSubnetCIDR                string
		expectedPrivateSubnetName                string
		expectedPrivateSubnetMapPublicIPOnLaunch bool
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "eu-central-1a",
						VPC: v1alpha1.AWSConfigSpecAWSVPC{
							PrivateSubnetCIDR: "10.1.2.0/25",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
			},
			expectedError:                            false,
			expectedPrivateSubnetAZ:                  "eu-central-1a",
			expectedPrivateSubnetCIDR:                "10.1.2.0/25",
			expectedPrivateSubnetName:                "test-cluster-private",
			expectedPrivateSubnetMapPublicIPOnLaunch: false,
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getSubnets(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.PrivateSubnetAZ != tc.expectedPrivateSubnetAZ {
				t.Errorf("unexpected PrivateSubnetAZ, got %q, want %q", a.PrivateSubnetAZ, tc.expectedPrivateSubnetAZ)
			}

			if a.PrivateSubnetCIDR != tc.expectedPrivateSubnetCIDR {
				t.Errorf("unexpected PrivateSubnetCIDR, got %q, want %q", a.PrivateSubnetCIDR, tc.expectedPrivateSubnetCIDR)
			}

			if a.PrivateSubnetName != tc.expectedPrivateSubnetName {
				t.Errorf("unexpected PrivateSubnetName, got %q, want %q", a.PrivateSubnetName, tc.expectedPrivateSubnetName)
			}

			if a.PrivateSubnetMapPublicIPOnLaunch != tc.expectedPrivateSubnetMapPublicIPOnLaunch {
				t.Errorf("unexpected PrivateSubnetMapPublicIPOnLaunch, got %t, want %t", a.PrivateSubnetMapPublicIPOnLaunch, tc.expectedPrivateSubnetMapPublicIPOnLaunch)
			}
		})
	}
}
