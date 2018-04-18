package adapter

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v10/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v10/templates/cloudformation/guest/route_tables.go
//

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
