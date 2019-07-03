package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type GuestVPCAdapter struct {
	CidrBlock        string
	ClusterID        string
	InstallationName string
	HostAccountID    string
	PeerVPCID        string
	PeerRoleArn      string
	Region           string
	RegionARN        string
	RouteTableNames  []RouteTableName
}

func (v *GuestVPCAdapter) Adapt(cfg Config) error {
	v.CidrBlock = key.StatusClusterNetworkCIDR(cfg.CustomObject)
	v.ClusterID = key.ClusterID(cfg.CustomObject)
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.ControlPlaneAccountID
	v.PeerVPCID = cfg.ControlPlaneVPCID
	v.Region = key.Region(cfg.CustomObject)
	v.RegionARN = key.RegionARN(cfg.CustomObject)
	v.PeerRoleArn = cfg.ControlPlanePeerRoleARN

	workerAZs := key.SortedWorkerAvailabilityZones(cfg.MachineDeployment)

	PublicRouteTable := RouteTableName{
		ResourceName: key.PublicRouteTableName(workerAZs[0]),
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic, workerAZs[0]),
	}
	v.RouteTableNames = append(v.RouteTableNames, PublicRouteTable)

	for _, az := range workerAZs {
		rtName := RouteTableName{
			ResourceName:        key.PrivateRouteTableName(az),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, az),
			VPCPeeringRouteName: key.VPCPeeringRouteName(az),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	return nil
}
