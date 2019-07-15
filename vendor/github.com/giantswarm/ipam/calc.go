package ipam

import "net"

// CalculateParent takes network as an input and returns one with 1 bit smaller
// mask (yielding therefore 1 bit larger network).
func CalculateParent(n net.IPNet) net.IPNet {
	ones, bits := n.Mask.Size()

	if ones > 0 {
		ones -= 1
	}

	n.Mask = net.CIDRMask(ones, bits)

	// Calculate network IP with new mask.
	n.IP = n.IP.Mask(n.Mask)

	return n
}

// Filter is basic functional filter function which iterates over given
// networks and returns ones that yield true with given filter function.
func Filter(networks []net.IPNet, f func(n net.IPNet) bool) []net.IPNet {
	var res []net.IPNet
	for _, n := range networks {
		if f(n) {
			res = append(res, n)
		}
	}
	return res
}
