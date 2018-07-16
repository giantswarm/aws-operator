package network

type Network struct {
	// CIDR is network segment from which IPAM allocates subnets for guest
	// clusters.
	CIDR string

	// PrivateSubnetMaskBits is number of bits in guest cluster private subnet
	// mask.
	PrivateSubnetMaskBits string

	// PublicSubnetMaskBits is number of bits in guest cluster public subnet
	// mask.
	PublicSubnetMaskBits string
}
