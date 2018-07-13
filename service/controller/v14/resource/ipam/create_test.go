package ipam

import (
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

func Test_deduplicateSubnets(t *testing.T) {
	testCases := []struct {
		name            string
		subnets         []net.IPNet
		expectedSubnets []net.IPNet
	}{
		{
			name:            "case 0: deduplicate empty list of subnets",
			subnets:         []net.IPNet{},
			expectedSubnets: []net.IPNet{},
		},
		{
			name: "case 1: deduplicate list of subnets with one element",
			subnets: []net.IPNet{
				mustParseCIDR("192.168.2.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.2.0/24"),
			},
		},
		{
			name: "case 2: deduplicate list of subnets with two non-overlapping elements",
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
			name: "case 3: deduplicate list of subnets with two overlapping elements",
			subnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.1.0/24"),
			},
			expectedSubnets: []net.IPNet{
				mustParseCIDR("192.168.1.0/24"),
			},
		},
		{
			name: "case 4: deduplicate list of subnets with four elements where two overlap",
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
			name: "case 5: same as case 4 but with different order",
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
			name: "case 6: same as case 4 but with different order",
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
			name: "case 7: same as case 4 but with different order",
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
			name: "case 7: deduplicate list of subnets with fiveelements where two overlap",
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subnets := deduplicateSubnets(tc.subnets)

			if !reflect.DeepEqual(subnets, tc.expectedSubnets) {
				t.Fatalf("expected %#v, got %#v", tc.expectedSubnets, subnets)
			}
		})
	}
}
