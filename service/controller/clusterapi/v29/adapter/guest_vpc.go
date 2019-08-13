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
	v.ClusterID = key.ClusterID(&cfg.CustomObject)
	v.InstallationName = cfg.InstallationName
	v.HostAccountID = cfg.ControlPlaneAccountID
	v.PeerVPCID = cfg.ControlPlaneVPCID
	v.Region = key.Region(cfg.CustomObject)
	v.RegionARN = key.RegionARN(cfg.AWSRegion)
	v.PeerRoleArn = cfg.ControlPlanePeerRoleARN

	for _, az := range cfg.TenantClusterAvailabilityZones {
		rtName := RouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PublicRouteTableName(az.Name)),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	for _, az := range cfg.TenantClusterAvailabilityZones {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		rtName := RouteTableName{
			ResourceName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
		}
		v.RouteTableNames = append(v.RouteTableNames, rtName)
	}

	return nil
}
