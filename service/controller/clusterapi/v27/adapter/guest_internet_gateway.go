package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
)

type GuestInternetGatewayAdapter struct {
	ClusterID          string
	PrivateRouteTables []string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = legacykey.ClusterID(cfg.CustomObject)

	for i := 0; i < len(legacykey.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		a.PrivateRouteTables = append(a.PrivateRouteTables, legacykey.PrivateRouteTableName(i))
	}

	return nil
}
