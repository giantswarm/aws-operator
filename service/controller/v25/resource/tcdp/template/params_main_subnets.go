package template

type ParamsMainSubnets []ParamsMainSubnetsSubnet

type ParamsMainSubnetsSubnet struct {
	AvailabilityZone      string
	CIDR                  string
	Name                  string
	RouteTableAssociation ParamsMainSubnetsSubnetRouteTableAssociation
	TCCP                  ParamsMainSubnetsSubnetTCCP
}

type ParamsMainSubnetsSubnetRouteTableAssociation struct {
	Name string
}

type ParamsMainSubnetsSubnetTCCP struct {
	Subnet ParamsMainSubnetsSubnetTCCPSubnet
	VPC    ParamsMainSubnetsSubnetTCCPVPC
}

type ParamsMainSubnetsSubnetTCCPSubnet struct {
	ID         string
	RouteTable ParamsMainSubnetsSubnetTCCPSubnetRouteTable
}

type ParamsMainSubnetsSubnetTCCPSubnetRouteTable struct {
	ID string
}

type ParamsMainSubnetsSubnetTCCPVPC struct {
	ID string
}
