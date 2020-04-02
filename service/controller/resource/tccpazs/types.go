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

func (m mapping) PublicSubnetEmpty() bool {
	return m.Public.Subnet.CIDR.IP == nil && m.Public.Subnet.CIDR.Mask == nil
}

func (m mapping) PrivateSubnetEmpty() bool {
	return m.Private.Subnet.CIDR.IP == nil && m.Private.Subnet.CIDR.Mask == nil
}

func (m mapping) AWSCNISubnetEmpty() bool {
	return m.AWSCNI.Subnet.CIDR.IP == nil && m.AWSCNI.Subnet.CIDR.Mask == nil
}
