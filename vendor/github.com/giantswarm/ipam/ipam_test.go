package ipam

import (
	"bytes"
	"net"
	"reflect"
	"testing"
)

// ipNetEqual returns true if the given IPNets refer to the same network.
func ipNetEqual(a, b net.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

// ipRangesEqual returns true if both given ipRanges are equal, false otherwise.
func ipRangesEqual(a, b []ipRange) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if !a[i].start.Equal(b[i].start) {
			return false
		}

		if !a[i].end.Equal(b[i].end) {
			return false
		}
	}

	return true
}

// TestIPToDecimal tests the ipToDecimal function.
func TestIPToDecimal(t *testing.T) {
	tests := []struct {
		ip              string
		expectedDecimal int
	}{
		{
			ip:              "0.0.0.0",
			expectedDecimal: 0,
		},
		{
			ip:              "0.0.05.3",
			expectedDecimal: 1283,
		},
		{
			ip:              "0.0.5.3",
			expectedDecimal: 1283,
		},
		{
			ip:              "10.0.0.0",
			expectedDecimal: 167772160,
		},
		{
			ip:              "10.4.0.0",
			expectedDecimal: 168034304,
		},
		{
			ip:              "255.255.255.255",
			expectedDecimal: 4294967295,
		},
	}

	for index, test := range tests {
		returnedDecimal := ipToDecimal(net.ParseIP(test.ip))

		if returnedDecimal != test.expectedDecimal {
			t.Fatalf(
				"%v: unexpected decimal returned.\nexpected: %v, returned: %v",
				index,
				test.expectedDecimal,
				returnedDecimal,
			)
		}
	}
}

// TestDecimalToIP tests the decimalToIP function.
func TestDecimalToIP(t *testing.T) {
	tests := []struct {
		decimal    int
		expectedIP string
	}{
		{
			decimal:    0,
			expectedIP: "0.0.0.0",
		},
		{
			decimal:    1283,
			expectedIP: "0.0.5.3",
		},
		{
			decimal:    167772160,
			expectedIP: "10.0.0.0",
		},
		{
			decimal:    168034304,
			expectedIP: "10.4.0.0",
		},
		{
			decimal:    4294967295,
			expectedIP: "255.255.255.255",
		},
	}

	for index, test := range tests {
		returnedIP := decimalToIP(test.decimal)
		expectedIP := net.ParseIP(test.expectedIP)

		if !returnedIP.Equal(expectedIP) {
			t.Fatalf(
				"%v: unexpected decimal returned.\nexpected: %v, returned: %v",
				index,
				expectedIP,
				returnedIP,
			)
		}
	}
}

// TestAdd tests the add function.
func TestAdd(t *testing.T) {
	tests := []struct {
		ip         string
		number     int
		expectedIP string
	}{
		{
			ip:         "127.0.0.1",
			number:     0,
			expectedIP: "127.0.0.1",
		},

		{
			ip:         "127.0.0.1",
			number:     1,
			expectedIP: "127.0.0.2",
		},

		{
			ip:         "127.0.0.1",
			number:     2,
			expectedIP: "127.0.0.3",
		},

		{
			ip:         "127.0.0.1",
			number:     -1,
			expectedIP: "127.0.0.0",
		},

		{
			ip:         "0.0.0.0",
			number:     -1,
			expectedIP: "255.255.255.255",
		},

		{
			ip:         "255.255.255.255",
			number:     1,
			expectedIP: "0.0.0.0",
		},
	}

	for index, test := range tests {
		ip := net.ParseIP(test.ip)
		expectedIP := net.ParseIP(test.expectedIP)

		returnedIP := add(ip, test.number)

		if !returnedIP.Equal(expectedIP) {
			t.Fatalf(
				"%v: unexpected ip returned.\nexpected: %v, returned: %v",
				index,
				expectedIP,
				returnedIP,
			)
		}
	}
}

// TestSize tests the Size function.
func TestSize(t *testing.T) {
	tests := []struct {
		mask         int
		expectedSize int
	}{
		{
			mask:         0,
			expectedSize: 4294967296,
		},
		{
			mask:         1,
			expectedSize: 2147483648,
		},
		{
			mask:         23,
			expectedSize: 512,
		},
		{
			mask:         24,
			expectedSize: 256,
		},
		{
			mask:         25,
			expectedSize: 128,
		},
		{
			mask:         31,
			expectedSize: 2,
		},
		{
			mask:         32,
			expectedSize: 1,
		},
	}

	for index, test := range tests {
		returnedSize := size(net.CIDRMask(test.mask, 32))

		if returnedSize != test.expectedSize {
			t.Fatalf(
				"%v: unexpected size returned.\nexpected: %v, returned: %v",
				index,
				test.expectedSize,
				returnedSize,
			)
		}
	}
}

// TestNewIPRange tests the newIPRange function.
func TestNewIPRange(t *testing.T) {
	tests := []struct {
		network         string
		expectedIPRange ipRange
	}{
		{
			network: "0.0.0.0/0",
			expectedIPRange: ipRange{
				start: net.ParseIP("0.0.0.0").To4(),
				end:   net.ParseIP("255.255.255.255").To4(),
			},
		},

		{
			network: "10.4.0.0/8",
			expectedIPRange: ipRange{
				start: net.ParseIP("10.0.0.0").To4(),
				end:   net.ParseIP("10.255.255.255").To4(),
			},
		},

		{
			network: "10.4.0.0/16",
			expectedIPRange: ipRange{
				start: net.ParseIP("10.4.0.0").To4(),
				end:   net.ParseIP("10.4.255.255").To4(),
			},
		},

		{
			network: "10.4.0.0/24",
			expectedIPRange: ipRange{
				start: net.ParseIP("10.4.0.0").To4(),
				end:   net.ParseIP("10.4.0.255").To4(),
			},
		},

		{
			network: "172.168.0.0/25",
			expectedIPRange: ipRange{
				start: net.ParseIP("172.168.0.0").To4(),
				end:   net.ParseIP("172.168.0.127").To4(),
			},
		},
	}

	for index, test := range tests {
		_, network, _ := net.ParseCIDR(test.network)
		ipRange := newIPRange(*network)

		if !reflect.DeepEqual(ipRange, test.expectedIPRange) {
			t.Fatalf(
				"%v: unexpected ipRange returned.\nexpected: %#v\nreturned: %#v\n",
				index,
				test.expectedIPRange,
				ipRange,
			)
		}
	}
}

// TestFreeIPRanges tests the freeIPRanges function.
func TestFreeIPRanges(t *testing.T) {
	tests := []struct {
		network              string
		subnets              []string
		expectedFreeIPRanges []ipRange
		expectedErrorHandler func(error) bool
	}{
		// Test that given a network with no subnets,
		// the entire network is returned as a free range.
		{
			network: "10.4.0.0/16",
			subnets: []string{},
			expectedFreeIPRanges: []ipRange{
				ipRange{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
		},

		// Test that given a network, and one subnet at the start of the network,
		// the entire remaining network - that is, the network minus the subnet,
		// is returned as a free range.
		{
			network: "10.4.0.0/16",
			subnets: []string{"10.4.0.0/24"},
			expectedFreeIPRanges: []ipRange{
				ipRange{
					start: net.ParseIP("10.4.1.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
		},

		// Test that given a network, and two contiguous subnets,
		// the entire remaining network (afterwards) is returned as a free range.
		{
			network: "10.4.0.0/16",
			subnets: []string{"10.4.0.0/24", "10.4.1.0/24"},
			expectedFreeIPRanges: []ipRange{
				ipRange{
					start: net.ParseIP("10.4.2.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
		},

		// Test that given a network, and one fragmented subnet,
		// the two remaining free ranges are returned as free.
		{
			network: "10.4.0.0/16",
			subnets: []string{"10.4.1.0/24"},
			expectedFreeIPRanges: []ipRange{
				ipRange{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.0.255"),
				},
				ipRange{
					start: net.ParseIP("10.4.2.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
		},

		// Test that given a network, and three fragmented subnets,
		// the 4 remaining free ranges are returned as free.
		{
			network: "10.4.0.0/16",
			subnets: []string{
				"10.4.10.0/24",
				"10.4.12.0/24",
				"10.4.14.0/24",
			},
			expectedFreeIPRanges: []ipRange{
				ipRange{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.9.255"),
				},
				ipRange{
					start: net.ParseIP("10.4.11.0"),
					end:   net.ParseIP("10.4.11.255"),
				},
				ipRange{
					start: net.ParseIP("10.4.13.0"),
					end:   net.ParseIP("10.4.13.255"),
				},
				ipRange{
					start: net.ParseIP("10.4.15.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
		},
	}

	for index, test := range tests {
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: could not parse cidr: %v", index, test.network)
		}

		subnets := []net.IPNet{}
		for _, subnetString := range test.subnets {
			_, subnet, err := net.ParseCIDR(subnetString)
			if err != nil {
				t.Fatalf("%v: could not parse cidr: %v", index, subnetString)
			}

			subnets = append(subnets, *subnet)
		}

		freeSubnets, err := freeIPRanges(*network, subnets)

		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned.\nreturned: %v", index, err)
			}
			if !test.expectedErrorHandler(err) {
				t.Fatalf("%v: incorrect error returned.\nreturned: %v", index, err)
			}
		} else {
			if test.expectedErrorHandler != nil {
				t.Fatalf("%v: expected error not returned.", index)
			}

			if !ipRangesEqual(freeSubnets, test.expectedFreeIPRanges) {
				t.Fatalf(
					"%v: unexpected free subnets returned.\nexpected: %v\nreturned: %v",
					index,
					test.expectedFreeIPRanges,
					freeSubnets,
				)
			}
		}
	}
}

// TestSpace tests the space function.
func TestSpace(t *testing.T) {
	tests := []struct {
		freeIPRanges         []ipRange
		mask                 int
		expectedIP           net.IP
		expectedErrorHandler func(error) bool
	}{
		// Test a case of fitting a network into an unused network.
		{
			freeIPRanges: []ipRange{
				{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
			mask:       24,
			expectedIP: net.ParseIP("10.4.0.0"),
		},

		// Test fitting a network into a network, with one subnet,
		// at the start of the range.
		{
			freeIPRanges: []ipRange{
				{
					start: net.ParseIP("10.4.1.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
			mask:       24,
			expectedIP: net.ParseIP("10.4.1.0"),
		},

		// Test adding a network that fills the range
		{
			freeIPRanges: []ipRange{
				{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
			mask:       16,
			expectedIP: net.ParseIP("10.4.0.0"),
		},

		// Test adding a network that is too large.
		{
			freeIPRanges: []ipRange{
				{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
			mask:                 15,
			expectedErrorHandler: IsSpaceExhausted,
		},

		// Test adding a slightly larger network,
		// given a smaller, non-contiguous subnet.
		{
			freeIPRanges: []ipRange{
				{
					start: net.ParseIP("10.4.0.0"),
					end:   net.ParseIP("10.4.0.255"),
				},
				{
					start: net.ParseIP("10.4.2.0"),
					end:   net.ParseIP("10.4.255.255"),
				},
			},
			mask:       23,
			expectedIP: net.ParseIP("10.4.2.0"),
		},
	}

	for index, test := range tests {
		mask := net.CIDRMask(test.mask, 32)

		ip, err := space(test.freeIPRanges, mask)

		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned.\nreturned: %v", index, err)
			}
			if !test.expectedErrorHandler(err) {
				t.Fatalf("%v: incorrect error returned.\nreturned: %v", index, err)
			}
		} else {
			if test.expectedErrorHandler != nil {
				t.Fatalf("%v: expected error not returned.", index)
			}

			if !ip.Equal(test.expectedIP) {
				t.Fatalf(
					"%v: unexpected ip returned. \nexpected: %v\nreturned: %v",
					index,
					test.expectedIP,
					ip,
				)
			}
		}
	}
}

// TestFree tests the Free function.
func TestFree(t *testing.T) {
	tests := []struct {
		network              string
		mask                 int
		subnets              []string
		expectedNetwork      string
		expectedErrorHandler func(error) bool
	}{
		// Test that a network with no existing subnets returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			subnets:         []string{},
			expectedNetwork: "10.4.0.0/24",
		},

		// Test that a network with one existing subnet returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			subnets:         []string{"10.4.0.0/24"},
			expectedNetwork: "10.4.1.0/24",
		},

		// Test that a network with two existing (non-fragmented) subnets returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			subnets:         []string{"10.4.0.0/24", "10.4.1.0/24"},
			expectedNetwork: "10.4.2.0/24",
		},

		// Test that a network with an existing subnet, that is fragmented,
		// and can fit one network before, returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			subnets:         []string{"10.4.1.0/24"},
			expectedNetwork: "10.4.0.0/24",
		},

		// Test that a network with an existing subnet, that is fragmented,
		// and can fit one network before, returns the correct subnet,
		// given a smaller mask.
		{
			network:         "10.4.0.0/16",
			mask:            25,
			subnets:         []string{"10.4.1.0/24"},
			expectedNetwork: "10.4.0.0/25",
		},

		// Test that a network with an existing subnet, that is fragmented,
		// but can't fit the requested network size before, returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            23,
			subnets:         []string{"10.4.1.0/24"}, // 10.4.1.0 - 10.4.1.255
			expectedNetwork: "10.4.2.0/23",           // 10.4.2.0 - 10.4.3.255
		},

		// Test that a network with no existing subnets returns the correct subnet,
		// for a mask that does not fall on an octet boundary.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			subnets:         []string{},
			expectedNetwork: "10.4.0.0/26",
		},

		// Test that a network with one existing subnet returns the correct subnet,
		// for a mask that does not fall on an octet boundary.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			subnets:         []string{"10.4.0.0/26"},
			expectedNetwork: "10.4.0.64/26",
		},

		// Test that a network with two existing fragmented subnets,
		// with a mask that does not fall on an octet boundary, returns the correct subnet.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			subnets:         []string{"10.4.0.0/26", "10.4.0.128/26"},
			expectedNetwork: "10.4.0.64/26",
		},

		// Test a setup with multiple, fragmented networks, of different sizes.
		{
			network: "10.4.0.0/24",
			mask:    29,
			subnets: []string{
				"10.4.0.0/26",
				"10.4.0.64/28",
				"10.4.0.80/28",
				"10.4.0.112/28",
				"10.4.0.128/26",
			},
			expectedNetwork: "10.4.0.96/29",
		},

		// Test where a network the same size as the main network is requested.
		{
			network:         "10.4.0.0/16",
			mask:            16,
			subnets:         []string{},
			expectedNetwork: "10.4.0.0/16",
		},

		// Test a setup where a network larger than the main network is requested.
		{
			network:              "10.4.0.0/16",
			mask:                 15,
			subnets:              []string{},
			expectedErrorHandler: IsMaskTooBig,
		},

		// Test where the existing networks are not ordered.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			subnets:         []string{"10.4.1.0/24", "10.4.0.0/24"},
			expectedNetwork: "10.4.2.0/24",
		},

		// Test where the existing networks are fragmented, and not ordered.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			subnets:         []string{"10.4.2.0/24", "10.4.0.0/24"},
			expectedNetwork: "10.4.1.0/24",
		},

		// Test where the range is full.
		{
			network:              "10.4.0.0/16",
			mask:                 17,
			subnets:              []string{"10.4.0.0/17", "10.4.128.0/17"},
			expectedErrorHandler: IsSpaceExhausted,
		},

		// Test where the subnet is not within the network.
		{
			network:              "10.4.0.0/16",
			mask:                 24,
			subnets:              []string{"10.5.0.0/24"},
			expectedErrorHandler: IsIPNotContained,
		},
	}

	for index, test := range tests {
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: could not parse cidr: %v", index, test.network)
		}

		mask := net.CIDRMask(test.mask, 32)

		subnets := []net.IPNet{}
		for _, e := range test.subnets {
			_, n, err := net.ParseCIDR(e)
			if err != nil {
				t.Fatalf("%v: could not parse cidr: %v", index, test.network)
			}
			subnets = append(subnets, *n)
		}

		_, expectedNetwork, _ := net.ParseCIDR(test.expectedNetwork)

		returnedNetwork, err := Free(*network, mask, subnets)

		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned.\nreturned: %v", index, err)
			}
			if !test.expectedErrorHandler(err) {
				t.Fatalf("%v: incorrect error returned.\nreturned: %v", index, err)
			}
		} else {
			if test.expectedErrorHandler != nil {
				t.Fatalf("%v: expected error not returned.", index)
			}

			if !ipNetEqual(returnedNetwork, *expectedNetwork) {
				t.Fatalf(
					"%v: unexpected network returned. \nexpected: %s (%#v, %#v) \nreturned: %s (%#v, %#v)",
					index,

					expectedNetwork.String(),
					expectedNetwork.IP,
					expectedNetwork.Mask,

					returnedNetwork.String(),
					returnedNetwork.IP,
					returnedNetwork.Mask,
				)
			}
		}
	}
}
