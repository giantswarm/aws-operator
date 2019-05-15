package ipam

import (
	"fmt"
	"net"
	"reflect"
	"testing"
)

func mustParseCIDR(val string) net.IPNet {
	_, n, err := net.ParseCIDR(val)
	if err != nil {
		panic(err)
	}

	return *n
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
