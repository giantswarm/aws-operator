package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type GuestInternetGatewayAdapter struct {
	ClusterID            string
	PublicRouteTableName string
	PrivateRouteTables   []string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(&cfg.CustomObject)

	a.PublicRouteTableName = key.SanitizeCFResourceName(key.PublicRouteTableName(key.MasterAvailabilityZone(cfg.CustomObject)))

	for _, az := range key.WorkerAvailabilityZones(cfg.MachineDeployment) {
		a.PrivateRouteTables = append(a.PrivateRouteTables, key.SanitizeCFResourceName(key.PrivateRouteTableName(az)))
	}

	return nil
}
