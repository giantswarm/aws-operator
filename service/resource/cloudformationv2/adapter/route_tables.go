package adapter

import (
	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/guest/route_tables.yaml

type routeTablesAdapter struct {
	PublicRouteTableName  string
	PrivateRouteTableName string
}

func (r *routeTablesAdapter) getRouteTables(cfg Config) error {
	r.PublicRouteTableName = keyv2.RouteTableName(cfg.CustomObject, suffixPublic)
	r.PrivateRouteTableName = keyv2.RouteTableName(cfg.CustomObject, suffixPrivate)

	return nil
}
