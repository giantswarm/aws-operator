package ipam

import (
	"fmt"
	"net"
	"strings"
)

// encodeKey returns a full storage key for a given network.
// e.g: 10.4.0.0/16 -> /ipam/subnet/10.4.0.0-16
func encodeKey(network net.IPNet) string {
	return fmt.Sprintf(
		IPAMSubnetStorageKeyFormat,
		strings.Replace(network.String(), "/", "-", -1),
	)
}

// decodeRelativeKey returns a CIDR string, given a relative storage key.
// e.g: 10.4.0.0-16 -> 10.4.0.0/16
func decodeRelativeKey(key string) string {
	return strings.Replace(key, "-", "/", -1)
}
