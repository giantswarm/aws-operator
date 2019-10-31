package template

type ParamsMainRouteTables struct {
	List []ParamsMainRouteTablesListItem
}

type ParamsMainRouteTablesListItem struct {
	AvailabilityZone string
	Name             string
	Route            ParamsMainRouteTablesListItemRoute
	TCCP             ParamsMainRouteTablesListItemTCCP
}

type ParamsMainRouteTablesListItemRoute struct {
	Name string
}

type ParamsMainRouteTablesListItemTCCP struct {
	NATGateway ParamsMainRouteTablesListItemTCCPNATGateway
	VPC        ParamsMainRouteTablesListItemTCCPVPC
}

type ParamsMainRouteTablesListItemTCCPNATGateway struct {
	ID string
}
type ParamsMainRouteTablesListItemTCCPVPC struct {
	ID string
}
