package template

type ParamsMainRouteTables struct {
	PeeringConnections []ParamsMainVPCPeeringConnection
}

type ParamsMainVPCPeeringConnection struct {
	ID         string
	Name       string
	RouteTable ParamsMainVPCPeeringConnectionRouteTable
	Subnet     ParamsMainVPCPeeringConnectionSubnet
}

type ParamsMainVPCPeeringConnectionRouteTable struct {
	ID string
}

type ParamsMainVPCPeeringConnectionSubnet struct {
	CIDR string
}
