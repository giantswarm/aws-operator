package template

type ParamsMainSubnets struct {
	AWSCNISubnets []ParamsMainSubnetsSubnet
	PublicSubnets []ParamsMainSubnetsSubnet
}

type ParamsMainSubnetsSubnet struct {
	AvailabilityZone      string
	CIDR                  string
	Name                  string
	MapPublicIPOnLaunch   bool
	RouteTableAssociation ParamsMainSubnetsSubnetRouteTableAssociation
}

type ParamsMainSubnetsSubnetRouteTableAssociation struct {
	Name           string
	RouteTableName string
	SubnetName     string
}
