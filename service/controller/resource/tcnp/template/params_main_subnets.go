package template

type ParamsMainSubnets struct {
	List []ParamsMainSubnetsListItem
}

type ParamsMainSubnetsListItem struct {
	AvailabilityZone      string
	CIDR                  string
	Name                  string
	RouteTable            ParamsMainSubnetsListItemRouteTable
	RouteTableAssociation ParamsMainSubnetsListItemRouteTableAssociation
	TagInternalELB        bool
	TCCP                  ParamsMainSubnetsListItemTCCP
}

type ParamsMainSubnetsListItemRouteTable struct {
	Name string
}

type ParamsMainSubnetsListItemRouteTableAssociation struct {
	Name string
}

type ParamsMainSubnetsListItemTCCP struct {
	VPC ParamsMainSubnetsListItemTCCPVPC
}

type ParamsMainSubnetsListItemTCCPVPC struct {
	ID string
}
