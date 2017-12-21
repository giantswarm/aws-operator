package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

// template related to this adapter: service/templates/cloudformation/nat_gateway.yaml

type natGatewayAdapter struct {
	PublicSubnetID      string
	PrivateRouteTableID string
}

func (n *natGatewayAdapter) getNatGateway(customObject v1alpha1.AWSConfig, clients Clients) error {
	// subnet ID
	// TODO: remove this code once the subnet is created by cloudformation and add a
	// reference in the template
	subnetName := keyv2.SubnetName(customObject, suffixPublic)
	subnetID, err := SubnetID(clients, subnetName)
	if err != nil {
		return microerror.Mask(err)
	}
	n.PublicSubnetID = subnetID

	// private route table ID
	// TODO: remove this code once the subnet is created by cloudformation and add a
	// reference in the template
	routeTableName := keyv2.RouteTableName(customObject, suffixPrivate)
	routeTableID, err := RouteTableID(clients, routeTableName)
	if err != nil {
		return microerror.Mask(err)
	}
	n.PrivateRouteTableID = routeTableID

	return nil
}
