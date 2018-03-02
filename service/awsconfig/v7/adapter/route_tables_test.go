package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterRouteTablesRegularFields(t *testing.T) {
	testCases := []struct {
		description                   string
		customObject                  v1alpha1.AWSConfig
		expectedError                 bool
		expectedHostClusterCIDR       string
		expectedPublicRouteTableName  string
		expectedPrivateRouteTableName string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
			},
			expectedError:                 false,
			expectedHostClusterCIDR:       "10.0.0.0/16",
			expectedPublicRouteTableName:  "test-cluster-public",
			expectedPrivateRouteTableName: "test-cluster-private",
		},
	}

	for _, tc := range testCases {
		hostClients := Clients{
			EC2: &EC2ClientMock{
				vpcCIDR: tc.expectedHostClusterCIDR,
			},
		}

		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      Clients{},
				HostClients:  hostClients,
			}
			err := a.getRouteTables(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.HostClusterCIDR != tc.expectedHostClusterCIDR {
				t.Errorf("unexpected HostClusterCIDR, got %q, want %q", a.HostClusterCIDR, tc.expectedHostClusterCIDR)
			}

			if a.PublicRouteTableName != tc.expectedPublicRouteTableName {
				t.Errorf("unexpected PublicRouteTableName, got %q, want %q", a.PublicRouteTableName, tc.expectedPrivateRouteTableName)
			}

			if a.PrivateRouteTableName != tc.expectedPrivateRouteTableName {
				t.Errorf("unexpected PrivateRouteTableName, got %q, want %q", a.PrivateRouteTableName, tc.expectedPrivateRouteTableName)
			}
		})
	}
}
