package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterInternetGatewayVPCID(t *testing.T) {
	testCases := []struct {
		description   string
		customObject  v1alpha1.AWSConfig
		expectedVPCID string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
				},
			},
			expectedVPCID: "vpc-12345",
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
				vpcID: tc.expectedVPCID,
			}
			err := a.getInternetGateway(tc.customObject, clients)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.VPCID != tc.expectedVPCID {
				t.Errorf("unexpected VPCID, got %q, want %q", a.VPCID, tc.expectedVPCID)
			}
		})
	}
}
