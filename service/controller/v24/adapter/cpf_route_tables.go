package adapter

type CPFRouteTables struct {
	PrivateRoutes []CPFRouteTablesRoute
	PublicRoutes  []CPFRouteTablesRoute
}

type CPFRouteTablesRoute struct {
	RouteTableName   string
	RouteTableID     string
	CidrBlock        string
	PeerConnectionID string
}
