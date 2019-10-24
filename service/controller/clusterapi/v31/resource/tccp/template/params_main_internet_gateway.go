package template

type GuestInternetGateway struct {
	ClusterID        string
	InternetGateways []ParamsInternetGatewayInternetGateway
}

type ParamsInternetGatewayInternetGateway struct {
	InternetGatewayRoute string
	RouteTable           string
}
