package aws

type VPC struct {
	CIDR              string `json:"cidr" yaml:"cidr"`
	PrivateSubnetCIDR string `json:"privateSubnetCidr" yaml:"privateSubnetCidr"`
	PublicSubnetCIDR  string `json:"publicSubnetCidr" yaml:"publicSubnetCidr"`
}
