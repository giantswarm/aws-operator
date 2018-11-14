package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v19/key"
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
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, key.PeerID(cfg.CustomObject))
	if err != nil {
		return microerror.Mask(err)
	}

	r.HostClusterCIDR = hostClusterCIDR
	r.PublicRouteTableName = RouteTableName{
		ResourceName: "PublicRouteTable00",
		TagName:      key.RouteTableName(cfg.CustomObject, suffixPublic),
	}

	for i := 0; i < key.SpecAvailabilityZones(cfg.CustomObject); i++ {
		suffix := fmt.Sprintf("%s%02d", suffixPrivate, i)
		rtName := RouteTableName{
			ResourceName:        fmt.Sprintf("PrivateRouteTable%02d", i),
			TagName:             key.RouteTableName(cfg.CustomObject, suffix),
			VPCPeeringRouteName: fmt.Sprintf("VPCPeeringRoute%02d", i),
		}
		r.PrivateRouteTableNames = append(r.PrivateRouteTableNames, rtName)
	}

	return nil
}
