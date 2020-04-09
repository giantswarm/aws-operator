package template

type ParamsMainRouteTables struct {
	PrivateRoutes []ParamsMainRouteTablesRoute
}

type ParamsMainRouteTablesRoute struct {
	RouteTableID     string
	CidrBlock        string
	PeerConnectionID string
}
