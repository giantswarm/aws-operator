package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type RouteTableName struct {
	ResourceName        string
	TagName             string
	VPCPeeringRouteName string
}

type GuestRouteTablesAdapter struct {
	HostClusterCIDR        string
	PublicRouteTableName   RouteTableName
	PrivateRouteTableNames []RouteTableName
}

func (r *GuestRouteTablesAdapter) Adapt(cfg Config) error {
	workerAZs := key.SortedWorkerAvailabilityZones(cfg.MachineDeployment)

	r.HostClusterCIDR = cfg.ControlPlaneVPCCidr
	r.PublicRouteTableName = RouteTableName{
		ResourceName: "PublicRouteTable",
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic, workerAZs[0]),
	}

	for _, az := range workerAZs {
		rtName := RouteTableName{
			ResourceName:        key.PrivateRouteTableName(az),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, az),
			VPCPeeringRouteName: key.VPCPeeringRouteName(az),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
