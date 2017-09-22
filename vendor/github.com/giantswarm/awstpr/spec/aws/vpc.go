package aws

type VPC struct {
	CIDR              string `json:"cidr" yaml:"cidr"`
	PrivateSubnetCIDR string `json:"privateSubnetCidr" yaml:"privateSubnetCidr"`
	PublicSubnetCIDR  string `json:"publicSubnetCidr" yaml:"publicSubnetCidr"`

	// RouteTableNames are the worker route tables of the master.
	// They are needed for a route to guest VPC in private host cluster subnets.
	RouteTableNames []string `json:"routeTableNames" yaml:"routeTableNames"`

	// PeerID is the ID of the VPC which we should peer to.
	// e.g: the vpc of the host cluster.
	PeerID string `json:"peerId" yaml:"peerId"`
}
