package adapter

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/guest/route_tables.yaml

type routeTablesAdapter struct {
	HostClusterCIDR       string
	PublicRouteTableName  string
	PrivateRouteTableName string
}

func (r *routeTablesAdapter) getRouteTables(cfg Config) error {
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, keyv2.PeerID(cfg.CustomObject))
	if err != nil {
		return microerror.Mask(err)
	}

	r.HostClusterCIDR = hostClusterCIDR
	r.PublicRouteTableName = keyv2.RouteTableName(cfg.CustomObject, suffixPublic)
	r.PrivateRouteTableName = keyv2.RouteTableName(cfg.CustomObject, suffixPrivate)

	return nil
}
