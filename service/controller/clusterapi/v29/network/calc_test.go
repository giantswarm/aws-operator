package network

import (
	"net"
	"reflect"
	"testing"
)

func Test_CalculateParent(t *testing.T) {
	testCases := []struct {
		name     string
		input    net.IPNet
		expected net.IPNet
	}{
		{
			name:     "case 0: calculate parent of 192.168.3.0/24",
			input:    mustParseCIDR("192.168.3.0/24"),
			expected: mustParseCIDR("192.168.2.0/23"),
		},
		{
			name:     "case 1: calculate parent of 10.100.3.96/27",
			input:    mustParseCIDR("10.100.3.96/27"),
			expected: mustParseCIDR("10.100.3.64/26"),
		},
		{
			name:     "case 2: calculate parent of 0.0.0.0/0",
			input:    mustParseCIDR("0.0.0.0/0"),
			expected: mustParseCIDR("0.0.0.0/0"),
		},
		{
			name:     "case 3: calculate parent of 255.255.255.255/32",
			input:    mustParseCIDR("255.255.255.255/32"),
			expected: mustParseCIDR("255.255.255.254/31"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := CalculateParent(tc.input)

			if !reflect.DeepEqual(output, tc.expected) {
				t.Fatalf("got %q, want %q", output, tc.expected)
			}
		})
	}
}

func Test_Filter(t *testing.T) {
	testCases := []struct {
		name       string
		input      []net.IPNet
		filterFunc func(net.IPNet) bool
		expected   []net.IPNet
	}{
		{
			name: "case 0: filter out specific network",
			input: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.4.0/24"),
			},
			filterFunc: func(n net.IPNet) bool {
				return !reflect.DeepEqual(mustParseCIDR("192.168.2.0/24"), n)
			},
			expected: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.3.0/24"),
				mustParseCIDR("192.168.4.0/24"),
			},
		},
		{
			name: "case 1: filter out by mask",
			input: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.8/24"),
				mustParseCIDR("192.168.3.33/25"),
				mustParseCIDR("192.168.3.136/27"),
				mustParseCIDR("192.168.4.0/24"),
			},
			filterFunc: func(n net.IPNet) bool {
				filter := mustParseCIDR("192.168.2.0/23")
				netIP := n.IP.Mask(filter.Mask)
				return !reflect.DeepEqual(filter.IP, netIP)
			},
			expected: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.4.0/24"),
			},
		},
		{
			name: "case 2: filter out everything",
			input: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.8/24"),
				mustParseCIDR("192.168.3.33/25"),
				mustParseCIDR("192.168.3.136/27"),
				mustParseCIDR("192.168.4.0/24"),
			},
			filterFunc: func(n net.IPNet) bool {
				return false
			},
			expected: []net.IPNet{},
		},
		{
			name: "case 3: filter out nothing",
			input: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.8/24"),
				mustParseCIDR("192.168.3.33/25"),
				mustParseCIDR("192.168.3.136/27"),
				mustParseCIDR("192.168.4.0/24"),
			},
			filterFunc: func(n net.IPNet) bool {
				return true
			},
			expected: []net.IPNet{
				mustParseCIDR("192.168.0.0/24"),
				mustParseCIDR("192.168.1.0/24"),
				mustParseCIDR("192.168.2.8/24"),
				mustParseCIDR("192.168.3.33/25"),
				mustParseCIDR("192.168.3.136/27"),
				mustParseCIDR("192.168.4.0/24"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := Filter(tc.input, tc.filterFunc)

			if len(output) == 0 && len(tc.expected) == 0 {
				// DeepEqual doesn't work correctly on empty slice.
				return
			}

			if !reflect.DeepEqual(output, tc.expected) {
				t.Fatalf("got %q, want %q", output, tc.expected)
			}
		})
	}
}
