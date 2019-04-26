package ipam

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
}

func Test_selectRandomAZs_properties(t *testing.T) {
	testCases := []struct {
		name         string
		azs          []string
		n            int
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: select one AZ out of three",
			azs:          []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
			n:            1,
			errorMatcher: nil,
		},
		{
			name:         "case 1: select three AZs out of three",
			azs:          []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
			n:            2,
			errorMatcher: nil,
		},
		{
			name:         "case 2: select three AZs out of three",
			azs:          []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
			n:            3,
			errorMatcher: nil,
		},
		{
			name:         "case 3: error when requesting more AZs than there are configured",
			azs:          []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
			n:            5,
			errorMatcher: IsInvalidParameter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Resource{
				logger:            microloggertest.New(),
				availabilityZones: tc.azs,
			}

			azs, err := r.selectRandomAZs(tc.n)

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

			if tc.errorMatcher != nil {
				return
			}

			if tc.n != len(azs) {
				t.Fatalf("got %d AZs in the first round, expected %d", len(azs), tc.n)
			}
		})
	}
}

func Test_selectRandomAZs_random(t *testing.T) {
	originalAZs := []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}
	r := Resource{
		logger:            microloggertest.New(),
		availabilityZones: originalAZs,
	}

	nTestRounds := 25
	numAZs := len(originalAZs) - 1
	selectedAZs := make([][]string, 0)

	for i := 0; i < nTestRounds; i++ {
		azs, err := r.selectRandomAZs(numAZs)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		differsFromOriginal := false
		differsFromEarlier := false
		for _, selectedAZs := range selectedAZs {
			for j, az := range originalAZs[:numAZs] {
				if azs[j] != az {
					differsFromOriginal = true
				}
			}

			for j, az := range selectedAZs {
				if azs[j] != az {
					differsFromEarlier = true
				}
			}

			if differsFromOriginal && differsFromEarlier {
				return
			}
		}

		selectedAZs = append(selectedAZs, azs)
	}

	t.Fatalf("after %d test rounds there was no difference in generated AZs over time and order of original AZs", nTestRounds)
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
