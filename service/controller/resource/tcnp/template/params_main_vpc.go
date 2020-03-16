package template

type ParamsMainVPC struct {
	Cluster     ParamsMainVPCCluster
	Region      ParamsMainVPCRegion
	RouteTables []ParamsMainVPCRouteTable
	TCCP        ParamsMainVPCTCCP
	TCNP        ParamsMainVPCTCNP
}

type ParamsMainVPCCluster struct {
	ID string
}

type ParamsMainVPCRegion struct {
	ARN  string
	Name string
}

type ParamsMainVPCRouteTable struct {
	ControlPlane  ParamsMainVPCRouteTableControlPlane
	Route         ParamsMainVPCRouteTableRoute
	RouteTable    ParamsMainVPCRouteTableRouteTable
	TenantCluster ParamsMainVPCRouteTableTenantCluster
}

type ParamsMainVPCRouteTableControlPlane struct {
	VPC ParamsMainVPCRouteTableControlPlaneVPC
}

type ParamsMainVPCRouteTableControlPlaneVPC struct {
	CIDR string
}

type ParamsMainVPCRouteTableRoute struct {
	Name string
}

type ParamsMainVPCRouteTableRouteTable struct {
	Name string
}

type ParamsMainVPCRouteTableTenantCluster struct {
	PeeringConnectionID string
}

type ParamsMainVPCTCCP struct {
	VPC ParamsMainVPCTCCPVPC
}

type ParamsMainVPCTCCPVPC struct {
	ID string
}

type ParamsMainVPCTCNP struct {
	CIDR string
}
