package adapter

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch2/key"
)

type GuestRouteTablesAdapter struct {
	HostClusterCIDR       string
	PublicRouteTableName  string
	PrivateRouteTableName string
}

func (r *GuestRouteTablesAdapter) Adapt(cfg Config) error {
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, key.PeerID(cfg.CustomObject))
	if err != nil {
		return microerror.Mask(err)
	}

	r.HostClusterCIDR = hostClusterCIDR
	r.PublicRouteTableName = key.RouteTableName(cfg.CustomObject, suffixPublic)
	r.PrivateRouteTableName = key.RouteTableName(cfg.CustomObject, suffixPrivate)

	return nil
}
