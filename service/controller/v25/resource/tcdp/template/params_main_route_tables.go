package template

import (
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

type ParamsMainRouteTableName struct {
	ResourceName        string
	TagName             string
	VPCPeeringRouteName string
}

type ParamsMainRouteTables struct {
	HostClusterCIDR        string
	PublicRouteTableName   ParamsMainRouteTableName
	PrivateRouteTableNames []ParamsMainRouteTableName
}

func (r *ParamsMainRouteTables) Adapt(cfg Config) error {
	r.HostClusterCIDR = cfg.ControlPlaneVPCCidr
	r.PublicRouteTableName = ParamsMainRouteTableName{
		ResourceName: "PublicRouteTable",
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic, 0),
	}

	for i := 0; i < len(key.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		rtName := ParamsMainRouteTableName{
			ResourceName:        key.PrivateRouteTableName(i),
			TagName:             key.RouteTableName(cfg.CustomObject, suffixPrivate, i),
			VPCPeeringRouteName: key.VPCPeeringRouteName(i),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
