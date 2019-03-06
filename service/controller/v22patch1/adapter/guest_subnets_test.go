package adapter

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterSubnetsRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                   string
		customObject           v1alpha1.AWSConfig
		expectedPublicSubnets  []Subnet
		expectedPrivateSubnets []Subnet
		errorMatcher           func(error) bool
	}{
		{
			name: "case 0: basic test that subnets are present for all three AZs",
			customObject: v1alpha1.AWSConfig{
				Status: v1alpha1.AWSConfigStatus{
					AWS: v1alpha1.AWSConfigStatusAWS{
						AvailabilityZones: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
							v1alpha1.AWSConfigStatusAWSAvailabilityZone{
								Name: "eu-west-1b",
								Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
									Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
										CIDR: "10.100.1.0/25",
									},
									Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
										CIDR: "10.100.1.128/25",
									},
								},
							},
							v1alpha1.AWSConfigStatusAWSAvailabilityZone{
								Name: "eu-west-1a",
								Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
									Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
										CIDR: "10.100.2.0/25",
									},
									Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
										CIDR: "10.100.2.128/25",
									},
								},
							},
							v1alpha1.AWSConfigStatusAWSAvailabilityZone{
								Name: "eu-west-1c",
								Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
									Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
										CIDR: "10.100.3.0/25",
									},
									Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
										CIDR: "10.100.3.128/25",
									},
								},
							},
						},
					},
				},
			},
			expectedPublicSubnets: []Subnet{
				{
					AvailabilityZone: "eu-west-1a",
					CIDR:             "10.100.2.0/25",
					Name:             "PublicSubnet",
					RouteTableAssociation: RouteTableAssociation{
						Name:           "PublicSubnetRouteTableAssociation",
						RouteTableName: "PublicRouteTable",
						SubnetName:     "PublicSubnet",
					},
				},
				{
					AvailabilityZone: "eu-west-1b",
					CIDR:             "10.100.1.0/25",
					Name:             "PublicSubnet01",
					RouteTableAssociation: RouteTableAssociation{
						Name:           "PublicSubnetRouteTableAssociation01",
						RouteTableName: "PublicRouteTable",
						SubnetName:     "PublicSubnet01",
					},
				},
				{
					AvailabilityZone: "eu-west-1c",
					CIDR:             "10.100.3.0/25",
					Name:             "PublicSubnet02",
					RouteTableAssociation: RouteTableAssociation{
						Name:           "PublicSubnetRouteTableAssociation02",
						RouteTableName: "PublicRouteTable",
						SubnetName:     "PublicSubnet02",
					},
				},
			},
			expectedPrivateSubnets: []Subnet{
				{
					AvailabilityZone: "eu-west-1a",
					CIDR:             "10.100.2.128/25",
					Name:             "PrivateSubnet",
					RouteTableAssociation: RouteTableAssociation{
						Name:           "PrivateSubnetRouteTableAssociation",
						RouteTableName: "PrivateRouteTable",
						SubnetName:     "PrivateSubnet",
					},
				},
				{
					AvailabilityZone: "eu-west-1b",
					CIDR:             "10.100.1.128/25",
					Name:             "PrivateSubnet01",
					RouteTableAssociation: RouteTableAssociation{
						Name:           "PrivateSubnetRouteTableAssociation01",
						RouteTableName: "PrivateRouteTable01",
						SubnetName:     "PrivateSubnet01",
					},
				},
				{
					AvailabilityZone: "eu-west-1c",
					CIDR:             "10.100.3.128/25",
					Name:             "PrivateSubnet02",
					RouteTableAssociation: RouteTableAssociation{
						Name:           "PrivateSubnetRouteTableAssociation02",
						RouteTableName: "PrivateRouteTable02",
						SubnetName:     "PrivateSubnet02",
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{}

		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}

			err := a.Guest.Subnets.Adapt(cfg)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(a.Guest.Subnets.PublicSubnets, tc.expectedPublicSubnets) {
				t.Fatalf("got PublicSubnets %#v, expected %#v", a.Guest.Subnets.PublicSubnets, tc.expectedPublicSubnets)
			}
			if !reflect.DeepEqual(a.Guest.Subnets.PrivateSubnets, tc.expectedPrivateSubnets) {
				t.Fatalf("got PrivateSubnets %#v, expected %#v", a.Guest.Subnets.PrivateSubnets, tc.expectedPrivateSubnets)
			}
		})
	}
}
