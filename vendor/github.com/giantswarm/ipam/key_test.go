package ipam

import (
	"net"
	"testing"
)

// TestEncodeKey tests the encodeKey function.
func TestEncodeKey(t *testing.T) {
	tests := []struct {
		network     string
		expectedKey string
	}{
		{
			network:     "10.4.0.0/16",
			expectedKey: "/ipam/subnet/10.4.0.0-16",
		},
		{
			network:     "192.168.1.0/24",
			expectedKey: "/ipam/subnet/192.168.1.0-24",
		},
	}

	for index, test := range tests {
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
		}

		returnedKey := encodeKey(*network)

		if returnedKey != test.expectedKey {
			t.Fatalf(
				"%v: returned key did not match expected key.\nexpected: %v\nreturned: %v\n",
				index,
				test.expectedKey,
				returnedKey,
			)
		}
	}
}

// TestDecodeRelativeKey tests the decodeRelativeKey function.
func TestDecodeRelativeKey(t *testing.T) {
	tests := []struct {
		key             string
		expectedNetwork string
	}{
		{
			key:             "10.4.0.0-16",
			expectedNetwork: "10.4.0.0/16",
		},
		{
			key:             "192.168.1.0-24",
			expectedNetwork: "192.168.1.0/24",
		},
	}

	for index, test := range tests {
		returnedNetwork := decodeRelativeKey(test.key)

		if returnedNetwork != test.expectedNetwork {
			t.Fatalf(
				"%v: returned network did not match expected network.\nexpected: %v\nreturned: %v\n",
				index,
				test.expectedNetwork,
				returnedNetwork,
			)
		}
	}
}
