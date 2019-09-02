package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/legacy/v30/key"
)

type GuestInternetGatewayAdapter struct {
	ClusterID          string
	PrivateRouteTables []string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(cfg.CustomObject)

	for i := 0; i < len(key.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		a.PrivateRouteTables = append(a.PrivateRouteTables, key.PrivateRouteTableName(i))
	}

	return nil
}
