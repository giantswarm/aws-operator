package ipam

import (
	"net"
)

// ipRange defines a pair of IPs, over a range.
type ipRange struct {
	start net.IP
	end   net.IP
}

// ipNets is a helper type for sorting net.IPNets.
type ipNets []net.IPNet

func (s ipNets) Len() int {
	return len(s)
}

func (s ipNets) Less(i, j int) bool {
	return ipToDecimal(s[i].IP) < ipToDecimal(s[j].IP)
}

func (s ipNets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
