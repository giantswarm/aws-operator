package template

type ParamsMainRouteTables struct {
	PrivateRoutes []ParamsMainRouteTablesRoute
	PublicRoutes  []ParamsMainRouteTablesRoute
}

type ParamsMainRouteTablesRoute struct {
	RouteTableName   string
	RouteTableID     string
	CidrBlock        string
	PeerConnectionID string
}
