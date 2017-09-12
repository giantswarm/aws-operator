package aws

type VPC struct {
	CIDR              string `json:"cidr" yaml:"cidr"`
	PrivateSubnetCIDR string `json:"privateSubnetCidr" yaml:"privateSubnetCidr"`
	PublicSubnetCIDR  string `json:"publicSubnetCidr" yaml:"publicSubnetCidr"`

	// PeerID is the ID of the VPC which we should peer to.
	// e.g: the vpc of the host cluster.
	PeerID string `json:"peerId" yaml:"peerId"`
}
