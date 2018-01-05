package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/route_tables.yaml

type routeTablesAdapter struct {
	PublicRouteTableName  string
	PrivateRouteTableName string
}

func (r *routeTablesAdapter) getRouteTables(customObject v1alpha1.AWSConfig, clients Clients) error {
	r.PublicRouteTableName = keyv2.RouteTableName(customObject, suffixPublic)
	r.PrivateRouteTableName = keyv2.RouteTableName(customObject, suffixPrivate)

	return nil
}
