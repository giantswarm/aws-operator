package adapter

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v4/key"
)

// template related to this adapter: service/templates/cloudformation/guest/route_tables.yaml

type routeTablesAdapter struct {
	HostClusterCIDR       string
	PublicRouteTableName  string
	PrivateRouteTableName string
}

func (r *routeTablesAdapter) getRouteTables(cfg Config) error {
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, key.PeerID(cfg.CustomObject))
	if err != nil {
		return microerror.Mask(err)
	}

	r.HostClusterCIDR = hostClusterCIDR
	r.PublicRouteTableName = key.RouteTableName(cfg.CustomObject, suffixPublic)
	r.PrivateRouteTableName = key.RouteTableName(cfg.CustomObject, suffixPrivate)

	return nil
}
