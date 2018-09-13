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
		"%s/%s",
		ipamSubnetStorageKey,
		strings.Replace(network.String(), "/", "-", -1),
	)
}

// decodeKey returns a CIDR string, given a storage key.
// e.g: /ipam/subnet/10.4.0.0-16 -> 10.4.0.0/16
func decodeKey(key string) string {
	key = strings.TrimPrefix(key, ipamSubnetStorageKey)
	key = strings.TrimPrefix(key, "/")
	return strings.Replace(key, "-", "/", -1)
}
