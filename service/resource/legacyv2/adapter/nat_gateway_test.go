package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterNATGatewaySubnetID(t *testing.T) {
	testCases := []struct {
		description            string
		customObject           v1alpha1.AWSConfig
		expectedPublicSubnetID string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			expectedPublicSubnetID: "subnet-12345",
		},
	}

	clients := Clients{
		KMS: &KMSClientMock{},
		IAM: &IAMClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			clients.EC2 = &EC2ClientMock{
				subnetID: tc.expectedPublicSubnetID,
			}
			err := a.getNatGateway(tc.customObject, clients)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.PublicSubnetID != tc.expectedPublicSubnetID {
				t.Errorf("unexpected PublicSubnetID, got %q, want %q", a.PublicSubnetID, tc.expectedPublicSubnetID)
			}
		})
	}
}
