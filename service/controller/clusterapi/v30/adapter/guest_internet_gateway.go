package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

type GuestInternetGatewayAdapter struct {
	ClusterID        string
	InternetGateways []GuestInternetGatewayAdapterInternetGateway
}

type GuestInternetGatewayAdapterInternetGateway struct {
	InternetGatewayRoute string
	RouteTable           string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(&cfg.CustomObject)

	for _, az := range cfg.TenantClusterAvailabilityZones {
		ig := GuestInternetGatewayAdapterInternetGateway{
			InternetGatewayRoute: key.SanitizeCFResourceName(key.PublicInternetGatewayRouteName(az.Name)),
			RouteTable:           key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}

		a.InternetGateways = append(a.InternetGateways, ig)
	}

	return nil
}
