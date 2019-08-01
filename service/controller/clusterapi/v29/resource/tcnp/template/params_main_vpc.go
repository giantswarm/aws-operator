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
	Name string
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
