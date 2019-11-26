package template

type ParamsMainInternetGateway struct {
	ClusterID        string
	InternetGateways []ParamsMainInternetGatewayInternetGateway
}

type ParamsMainInternetGatewayInternetGateway struct {
	InternetGatewayRoute string
	RouteTable           string
}
