package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterSubnetsRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description                              string
		customObject                             v1alpha1.AWSConfig
		expectedError                            bool
		expectedPublicSubnetAZ                   string
		expectedPublicSubnetCIDR                 string
		expectedPublicSubnetName                 string
		expectedPublicSubnetMapPublicIPOnLaunch  bool
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
							PublicSubnetCIDR:  "10.1.1.0/25",
							PrivateSubnetCIDR: "10.1.2.0/25",
						},
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
			},
			expectedError:                            false,
			expectedPublicSubnetAZ:                   "eu-central-1a",
			expectedPublicSubnetCIDR:                 "10.1.1.0/25",
			expectedPublicSubnetName:                 "test-cluster-public",
			expectedPublicSubnetMapPublicIPOnLaunch:  false,
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
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.Guest.Subnets.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.Subnets.PublicSubnetAZ != tc.expectedPublicSubnetAZ {
				t.Errorf("unexpected PublicSubnetAZ, got %q, want %q", a.Guest.Subnets.PublicSubnetAZ, tc.expectedPublicSubnetAZ)
			}

			if a.Guest.Subnets.PublicSubnetCIDR != tc.expectedPublicSubnetCIDR {
				t.Errorf("unexpected PublicSubnetCIDR, got %q, want %q", a.Guest.Subnets.PublicSubnetCIDR, tc.expectedPublicSubnetCIDR)
			}

			if a.Guest.Subnets.PublicSubnetName != tc.expectedPublicSubnetName {
				t.Errorf("unexpected PublicSubnetName, got %q, want %q", a.Guest.Subnets.PublicSubnetName, tc.expectedPublicSubnetName)
			}

			if a.Guest.Subnets.PublicSubnetMapPublicIPOnLaunch != tc.expectedPublicSubnetMapPublicIPOnLaunch {
				t.Errorf("unexpected PublicSubnetMapPublicIPOnLaunch, got %t, want %t", a.Guest.Subnets.PublicSubnetMapPublicIPOnLaunch, tc.expectedPublicSubnetMapPublicIPOnLaunch)
			}

			if a.Guest.Subnets.PrivateSubnetAZ != tc.expectedPrivateSubnetAZ {
				t.Errorf("unexpected PrivateSubnetAZ, got %q, want %q", a.Guest.Subnets.PrivateSubnetAZ, tc.expectedPrivateSubnetAZ)
			}

			if a.Guest.Subnets.PrivateSubnetCIDR != tc.expectedPrivateSubnetCIDR {
				t.Errorf("unexpected PrivateSubnetCIDR, got %q, want %q", a.Guest.Subnets.PrivateSubnetCIDR, tc.expectedPrivateSubnetCIDR)
			}

			if a.Guest.Subnets.PrivateSubnetName != tc.expectedPrivateSubnetName {
				t.Errorf("unexpected PrivateSubnetName, got %q, want %q", a.Guest.Subnets.PrivateSubnetName, tc.expectedPrivateSubnetName)
			}

			if a.Guest.Subnets.PrivateSubnetMapPublicIPOnLaunch != tc.expectedPrivateSubnetMapPublicIPOnLaunch {
				t.Errorf("unexpected PrivateSubnetMapPublicIPOnLaunch, got %t, want %t", a.Guest.Subnets.PrivateSubnetMapPublicIPOnLaunch, tc.expectedPrivateSubnetMapPublicIPOnLaunch)
			}
		})
	}
}
