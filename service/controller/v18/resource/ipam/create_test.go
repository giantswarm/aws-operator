package ipam

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
}

func Test_canonicalizeSubnets(t *testing.T) {
	testCases := []struct {
		name            string
		network         net.IPNet
		subnets         []net.IPNet
		expectedSubnets []net.IPNet
	}{
		{
			name:            "case 0: deduplicate empty list of subnets",
			network:         mustParseCIDR("192.168.0.0/16"),
			subnets:         []net.IPNet{},
			expectedSubnets: []net.IPNet{},
		},
		{
			name:    "case 1: deduplicate list of subnets with one element",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.2.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.2.0/24"),
			},
		},
		{
			name:    "case 2: deduplicate list of subnets with two non-overlapping elements",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
			},
		},
		{
			name:    "case 3: deduplicate list of subnets with two overlapping elements",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.1.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
			},
		},
		{
			name:    "case 4: deduplicate list of subnets with four elements where two overlap",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
		},
		{
			name:    "case 5: same as case 4 but with different order",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
		},
		{
			name:    "case 6: same as case 4 but with different order",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
		},
		{
			name:    "case 7: same as case 4 but with different order",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.1.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
			},
		},
		{
			name:    "case 7: deduplicate list of subnets with fiveelements where two overlap",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.4.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.4.0/24"),
			},
		},
		{
			name:    "case 8: deduplicate list of subnets with duplicates and IPs from different segments",
			network: mustParseCIDR("192.168.0.0/16"),
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.4.0/24"),
				mustParseCIDR("172.31.0.1/16"),
				mustParseCIDR("10.2.0.4/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.4.0/24"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subnets := canonicalizeSubnets(tc.network, tc.subnets)

			if !reflect.DeepEqual(subnets, tc.expectedSubnets) {
				msg := "expected: {\n"
				for _, n := range tc.expectedSubnets {
					msg += fmt.Sprintf("\t%s,\n", n.String())
				}
				msg += "}\n\ngot: {\n"
				for _, n := range subnets {
					msg += fmt.Sprintf("\t%s,\n", n.String())
				}
				msg += "}"
				t.Fatal(msg)
			}
		})
	}
}

func Test_splitSubnetToStatusAZs(t *testing.T) {
	testCases := []struct {
		name         string
		subnet       net.IPNet
		azs          []string
		expectedAZs  []v1alpha1.AWSConfigStatusAWSAvailabilityZone
		errorMatcher func(error) bool
	}{
		{
			name:   "case 0: split 10.100.4.0/22 for [eu-west-1b]",
			subnet: mustParseCIDR("10.100.4.0/22"),
			azs:    []string{"eu-west-1b"},
			expectedAZs: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
				v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					Name: "eu-west-1b",
					Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.0/23",
						},
						Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.6.0/23",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:   "case 1: split 10.100.4.0/22 for [eu-west-1c, eu-west-1a]",
			subnet: mustParseCIDR("10.100.4.0/22"),
			azs:    []string{"eu-west-1c", "eu-west-1a"},
			expectedAZs: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
				v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					Name: "eu-west-1c",
					Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.0/24",
						},
						Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.5.0/24",
						},
					},
				},
				v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					Name: "eu-west-1a",
					Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.6.0/24",
						},
						Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.7.0/24",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name:   "case 2: split 10.100.4.0/22 for three [eu-west-1a, eu-west-1b, eu-west-1c]",
			subnet: mustParseCIDR("10.100.4.0/22"),
			azs:    []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
			expectedAZs: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
				v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					Name: "eu-west-1a",
					Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.0/25",
						},
						Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.4.128/25",
						},
					},
				},
				v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					Name: "eu-west-1b",
					Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.5.0/25",
						},
						Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.5.128/25",
						},
					},
				},
				v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					Name: "eu-west-1c",
					Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.6.0/25",
						},
						Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.6.128/25",
						},
					},
				},
			},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			azs, err := splitSubnetToStatusAZs(tc.subnet, tc.azs)

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

			if !reflect.DeepEqual(azs, tc.expectedAZs) {
				t.Fatalf("got %q, expected %q", azs, tc.expectedAZs)
			}
		})
	}
}

func Test_calculateSubnetMask(t *testing.T) {
	testCases := []struct {
		name         string
		mask         net.IPMask
		n            uint
		expectedMask net.IPMask
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: split /24 into one network",
			mask:         net.CIDRMask(24, 32),
			n:            1,
			expectedMask: net.CIDRMask(24, 32),
			errorMatcher: nil,
		},
		{
			name:         "case 1: split /24 into two networks",
			mask:         net.CIDRMask(24, 32),
			n:            2,
			expectedMask: net.CIDRMask(25, 32),
			errorMatcher: nil,
		},
		{
			name:         "case 2: split /24 into three networks",
			mask:         net.CIDRMask(24, 32),
			n:            3,
			expectedMask: net.CIDRMask(26, 32),
			errorMatcher: nil,
		},
		{
			name:         "case 3: split /24 into four networks",
			mask:         net.CIDRMask(24, 32),
			n:            4,
			expectedMask: net.CIDRMask(26, 32),
			errorMatcher: nil,
		},
		{
			name:         "case 4: split /24 into five networks",
			mask:         net.CIDRMask(24, 32),
			n:            5,
			expectedMask: net.CIDRMask(27, 32),
			errorMatcher: nil,
		},
		{
			name:         "case 5: split /22 into 7 networks",
			mask:         net.CIDRMask(22, 32),
			n:            7,
			expectedMask: net.CIDRMask(25, 32),
			errorMatcher: nil,
		},
		{
			name:         "case 6: split /31 into 8 networks (no room)",
			mask:         net.CIDRMask(31, 32),
			n:            7,
			expectedMask: nil,
			errorMatcher: IsInvalidParameter,
		},
		{
			name:         "case 7: IPv6 masks (split /31 for seven networks)",
			mask:         net.CIDRMask(31, 128),
			n:            7,
			expectedMask: net.CIDRMask(34, 128),
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mask, err := calculateSubnetMask(tc.mask, tc.n)

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

			if !reflect.DeepEqual(mask, tc.expectedMask) {
				t.Fatalf("Mask == %q, want %q", mask, tc.expectedMask)
			}
		})
	}
}

func Test_splitNetwork(t *testing.T) {
	testCases := []struct {
		name            string
		network         net.IPNet
		n               uint
		expectedSubnets []net.IPNet
		errorMatcher    func(error) bool
	}{
		{
			name:    "case 0: split /24 into four networks",
			network: mustParseCIDR("192.168.8.0/24"),
			n:       4,
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.8.0/26"),
				mustParseCIDR("192.168.8.64/26"),
				mustParseCIDR("192.168.8.128/26"),
				mustParseCIDR("192.168.8.192/26"),
			},
			errorMatcher: nil,
		},
		{
			name:    "case 1: split /22 into 7 networks",
			network: mustParseCIDR("10.100.0.0/22"),
			n:       7,
			expectedSubnets: []net.IPNet{
				mustParseCIDR("10.100.0.0/25"),
				mustParseCIDR("10.100.0.128/25"),
				mustParseCIDR("10.100.1.0/25"),
				mustParseCIDR("10.100.1.128/25"),
				mustParseCIDR("10.100.2.0/25"),
				mustParseCIDR("10.100.2.128/25"),
				mustParseCIDR("10.100.3.0/25"),
			},
			errorMatcher: nil,
		},
		{
			name:            "case 2: split /31 into 8 networks (no room)",
			network:         mustParseCIDR("192.168.8.128/31"),
			n:               7,
			expectedSubnets: nil,
			errorMatcher:    IsInvalidParameter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subnets, err := splitNetwork(tc.network, tc.n)

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

			if !reflect.DeepEqual(subnets, tc.expectedSubnets) {
				msg := "expected: {\n"
				for _, n := range tc.expectedSubnets {
					msg += fmt.Sprintf("\t%s,\n", n.String())
				}
				msg += "}\n\ngot: {\n"
				for _, n := range subnets {
					msg += fmt.Sprintf("\t%s,\n", n.String())
				}
				msg += "}"
				t.Fatal(msg)

			}
		})
	}
}
