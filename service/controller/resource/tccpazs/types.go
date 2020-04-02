package tccpazs

import "net"

// mapping is temporary type for mapping existing subnets from controllercontext
// to AZs.
type mapping struct {
	AWSCNI       network
	Public       network
	Private      network
	RequiredByCR bool
}

type network struct {
	Subnet subnet
}

type subnet struct {
	CIDR net.IPNet
	ID   string
}

func (m mapping) subnetsEmpty() bool {
	return (m.Public.Subnet.CIDR.IP == nil && m.Public.Subnet.CIDR.Mask == nil) && (m.Private.Subnet.CIDR.IP == nil && m.Private.Subnet.CIDR.Mask == nil) && (m.AWSCNI.Subnet.CIDR.IP == nil && m.AWSCNI.Subnet.CIDR.Mask == nil)
}
