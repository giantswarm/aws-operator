package clusterazs

import "net"

// mapping is temporary type for mapping existing subnets from controllercontext
// to AZs.
type mapping struct {
	Public  network
	Private network
}

type network struct {
	RouteTable routetable
	Subnet     subnet
}

type routetable struct {
	ID string
}

type subnet struct {
	CIDR net.IPNet
	ID   string
}

func (m mapping) subnetsEmpty() bool {
	return (m.Public.Subnet.CIDR.IP == nil && m.Public.Subnet.CIDR.Mask == nil) && (m.Private.Subnet.CIDR.IP == nil && m.Private.Subnet.CIDR.Mask == nil)
}
