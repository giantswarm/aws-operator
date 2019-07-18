package clusterazs

import "net"

// subnetPair is temporary type for mapping existing subnets from
// controllercontext to AZs.
type subnetPair struct {
	// These members are exported so that go-cmp can make a diff for unit test
	// results.
	ID      string
	Public  net.IPNet
	Private net.IPNet
}

func (sp subnetPair) areEmpty() bool {
	return (sp.Public.IP == nil && sp.Public.Mask == nil) && (sp.Private.IP == nil && sp.Private.Mask == nil)
}
