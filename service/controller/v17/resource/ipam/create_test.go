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
