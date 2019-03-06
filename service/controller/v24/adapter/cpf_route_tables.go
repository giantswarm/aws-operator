package adapter

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/routetable"
)

type CPFRouteTablesConfig struct {
	RouteTable *routetable.RouteTable

	AvailabilityZones []v1alpha1.AWSConfigStatusAWSAvailabilityZone
	EncrypterBackend  string
	NetworkCIDR       string
}

type CPFRouteTables struct {
	routeTable *routetable.RouteTable

	PrivateRoutes []CPFRouteTablesRoute
	PublicRoutes  []CPFRouteTablesRoute

	availabilityZones []v1alpha1.AWSConfigStatusAWSAvailabilityZone
	encrypterBackend  string
	networkCIDR       string
}

func newCPFRouteTables(config CPFRouteTablesConfig) (*CPFRouteTables, error) {
	if config.RouteTable == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RouteTable must not be empty", config)
	}

	if config.AvailabilityZones == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AvailabilityZones must not be empty", config)
	}
	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}
	if config.NetworkCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkCIDR must not be empty", config)
	}

	r := &CPFRouteTables{
		routeTable: config.RouteTable,

		encrypterBackend:  config.EncrypterBackend,
		availabilityZones: config.AvailabilityZones,
		networkCIDR:       config.NetworkCIDR,
	}

	return r, nil
}

func (a *CPFRouteTables) Boot(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var tenantPrivateSubnetCidrs []string
	{
		for _, az := range a.availabilityZones {
			tenantPrivateSubnetCidrs = append(tenantPrivateSubnetCidrs, az.Subnet.Private.CIDR)
		}
	}

	// private routes.
	for _, name := range a.routeTable.Names() {
		id, err := a.routeTable.IDForName(ctx, name)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, cidrBlock := range tenantPrivateSubnetCidrs {
			route := CPFRouteTablesRoute{
				RouteTableName: name,
				RouteTableID:   id,
				// Requester CIDR block, we create the peering connection from the
				// guest's private subnets.
				CidrBlock: cidrBlock,
				// The peer connection id is fetched from the cloud formation stack
				// outputs in the stackoutput resource.
				PeerConnectionID: cc.Status.Cluster.VPCPeeringConnectionID,
			}

			a.PrivateRoutes = append(a.PrivateRoutes, route)
		}
	}

	// public routes for vault.
	if a.encrypterBackend == encrypter.VaultBackend {
		for _, name := range a.routeTable.Names() {
			id, err := a.routeTable.IDForName(ctx, name)
			if err != nil {
				return microerror.Mask(err)
			}

			route := CPFRouteTablesRoute{
				RouteTableName: name,
				RouteTableID:   id,
				// Requester CIDR block, we create the peering connection from the
				// guest's CIDR for being able to access Vault's ELB.
				CidrBlock: a.networkCIDR,
				// The peer connection id is fetched from the cloud formation stack
				// outputs in the stackoutput resource.
				PeerConnectionID: cc.Status.Cluster.VPCPeeringConnectionID,
			}

			a.PublicRoutes = append(a.PublicRoutes, route)
		}
	}

	return nil
}
