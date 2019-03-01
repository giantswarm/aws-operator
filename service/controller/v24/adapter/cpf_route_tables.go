package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

type CPFRouteTables struct {
	PrivateRoutes []CPFRouteTablesRoute
	PublicRoutes  []CPFRouteTablesRoute
}

func (a *CPFRouteTables) Adapt(ctx context.Context, config Config) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var tenantPrivateSubnetCidrs []string
	{
		for _, az := range key.StatusAvailabilityZones(config.CustomObject) {
			tenantPrivateSubnetCidrs = append(tenantPrivateSubnetCidrs, az.Subnet.Private.CIDR)
		}
	}

	// private routes.
	for _, routeTableName := range config.CustomObject.Spec.AWS.VPC.RouteTableNames {
		routeTableID, err := routeTableID(routeTableName, config)
		if err != nil {
			return microerror.Mask(err)
		}
		for _, cidrBlock := range tenantPrivateSubnetCidrs {
			rt := CPFRouteTablesRoute{
				RouteTableName: routeTableName,
				RouteTableID:   routeTableID,
				// Requester CIDR block, we create the peering connection from the guest's private subnets.
				CidrBlock:        cidrBlock,
				PeerConnectionID: cc.Status.Cluster.VPCPeeringConnectionID,
			}
			a.PrivateRoutes = append(a.PrivateRoutes, rt)
		}
	}

	// public routes for vault.
	if config.EncrypterBackend == encrypter.VaultBackend {
		publicRouteTables := strings.Split(config.PublicRouteTables, ",")
		for _, routeTableName := range publicRouteTables {
			routeTableID, err := routeTableID(routeTableName, config)
			if err != nil {
				return microerror.Mask(err)
			}
			rt := CPFRouteTablesRoute{
				RouteTableName: routeTableName,
				RouteTableID:   routeTableID,
				// Requester CIDR block, we create the peering connection from the
				// guest's CIDR for being able to access Vault's ELB.
				CidrBlock:        key.ClusterNetworkCIDR(config.CustomObject),
				PeerConnectionID: cc.Status.Cluster.VPCPeeringConnectionID,
			}
			a.PublicRoutes = append(a.PublicRoutes, rt)
		}
	}
	return nil
}

type CPFRouteTablesRoute struct {
	RouteTableName   string
	RouteTableID     string
	CidrBlock        string
	PeerConnectionID string
}

func routeTableID(name string, config Config) (string, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}
	output, err := config.HostClients.EC2.DescribeRouteTables(input)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(output.RouteTables) == 0 {
		return "", microerror.Maskf(tooFewResultsError, "route tables: %s", name)
	}
	if len(output.RouteTables) > 1 {
		return "", microerror.Maskf(tooManyResultsError, "route tables: %s", name)
	}

	return *output.RouteTables[0].RouteTableId, nil
}
