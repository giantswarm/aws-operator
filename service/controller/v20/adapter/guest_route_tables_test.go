package adapter

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterRouteTablesRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description                    string
		customObject                   v1alpha1.AWSConfig
		expectedError                  bool
		expectedHostClusterCIDR        string
		expectedPublicRouteTableName   RouteTableName
		expectedPrivateRouteTableNames []RouteTableName
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						AvailabilityZones: 2,
					},
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
				Status: v1alpha1.AWSConfigStatus{
					AWS: v1alpha1.AWSConfigStatusAWS{
						AvailabilityZones: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
							v1alpha1.AWSConfigStatusAWSAvailabilityZone{
								Name: "eu-central-1a",
							},
							v1alpha1.AWSConfigStatusAWSAvailabilityZone{
								Name: "eu-central-1b",
							},
						},
					},
				},
			},
			expectedError:           false,
			expectedHostClusterCIDR: "10.0.0.0/16",
			expectedPublicRouteTableName: RouteTableName{
				ResourceName: "PublicRouteTable",
				TagName:      "test-cluster-public",
			},
			expectedPrivateRouteTableNames: []RouteTableName{
				{
					ResourceName:        "PrivateRouteTable",
					TagName:             "test-cluster-private",
					VPCPeeringRouteName: "VPCPeeringRoute",
				},
				{
					ResourceName:        "PrivateRouteTable01",
					TagName:             "test-cluster-private01",
					VPCPeeringRouteName: "VPCPeeringRoute01",
				},
			},
		},
	}

	for _, tc := range testCases {
		hostClients := Clients{
			EC2: &EC2ClientMock{
				vpcCIDR: tc.expectedHostClusterCIDR,
			},
			STS: &STSClientMock{},
		}

		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      Clients{},
				HostClients:  hostClients,
			}
			err := a.Guest.RouteTables.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.RouteTables.HostClusterCIDR != tc.expectedHostClusterCIDR {
				t.Errorf("unexpected HostClusterCIDR, got %q, want %q", a.Guest.RouteTables.HostClusterCIDR, tc.expectedHostClusterCIDR)
			}

			if !reflect.DeepEqual(a.Guest.RouteTables.PublicRouteTableName, tc.expectedPublicRouteTableName) {
				t.Errorf("unexpected PublicRouteTableName, got %q, want %q", a.Guest.RouteTables.PublicRouteTableName, tc.expectedPublicRouteTableName)
			}

			if !reflect.DeepEqual(a.Guest.RouteTables.PrivateRouteTableNames, tc.expectedPrivateRouteTableNames) {
				t.Errorf("unexpected PrivateRouteTableNames, got %q, want %q", a.Guest.RouteTables.PrivateRouteTableNames, tc.expectedPrivateRouteTableNames)
			}
		})
	}
}
