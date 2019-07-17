package template

type ParamsMainSubnets struct {
	List []ParamsMainSubnetsListItem
}

type ParamsMainSubnetsListItem struct {
	AvailabilityZone      string
	CIDR                  string
	Name                  string
	RouteTableAssociation ParamsMainSubnetsListItemRouteTableAssociation
	TCCP                  ParamsMainSubnetsListItemTCCP
}

type ParamsMainSubnetsListItemRouteTableAssociation struct {
	Name string
}

type ParamsMainSubnetsListItemTCCP struct {
	Subnet ParamsMainSubnetsListItemTCCPSubnet
	VPC    ParamsMainSubnetsListItemTCCPVPC
}

type ParamsMainSubnetsListItemTCCPSubnet struct {
	ID         string
	RouteTable ParamsMainSubnetsListItemTCCPSubnetRouteTable
}

type ParamsMainSubnetsListItemTCCPSubnetRouteTable struct {
	ID string
}

type ParamsMainSubnetsListItemTCCPVPC struct {
	ID string
}
