package clusterazs

import "net"

// subnetPair is temporary type for mapping existing subnets from
// controllercontext to AZs.
type subnetPair struct {
	// These members are exported so that go-cmp can make a diff for unit test
	// results.
	Public  subnet
	Private subnet
}

type subnet struct {
	CIDR net.IPNet
	ID   string
}

func (sp subnetPair) areEmpty() bool {
	return (sp.Public.CIDR.IP == nil && sp.Public.CIDR.Mask == nil) && (sp.Private.CIDR.IP == nil && sp.Private.CIDR.Mask == nil)
}
