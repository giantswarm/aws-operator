package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v23/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v23/key"
)

type HostPostRouteTablesAdapter struct {
	PrivateRoutes []HostPostRouteTablesAdapterRoute
	PublicRoutes  []HostPostRouteTablesAdapterRoute
}

func (i *HostPostRouteTablesAdapter) Adapt(cfg Config) error {
	peerConnectionID, err := waitForPeeringConnectionID(cfg)
	if err != nil {
		return microerror.Mask(err)
	}

	var tenantPrivateSubnetCidrs []string
	{
		for _, az := range key.StatusAvailabilityZones(cfg.CustomObject) {
			tenantPrivateSubnetCidrs = append(tenantPrivateSubnetCidrs, az.Subnet.Private.CIDR)
		}
	}

	// private routes.
	for _, routeTableName := range cfg.CustomObject.Spec.AWS.VPC.RouteTableNames {
		routeTableID, err := routeTableID(routeTableName, cfg)
		if err != nil {
			return microerror.Mask(err)
		}
		for _, cidrBlock := range tenantPrivateSubnetCidrs {
			rt := HostPostRouteTablesAdapterRoute{
				RouteTableName: routeTableName,
				RouteTableID:   routeTableID,
				// Requester CIDR block, we create the peering connection from the guest's private subnets.
				CidrBlock:        cidrBlock,
				PeerConnectionID: peerConnectionID,
			}
			i.PrivateRoutes = append(i.PrivateRoutes, rt)
		}
	}

	// public routes for vault.
	if cfg.EncrypterBackend == encrypter.VaultBackend {
		publicRouteTables := strings.Split(cfg.PublicRouteTables, ",")
		for _, routeTableName := range publicRouteTables {
			routeTableID, err := routeTableID(routeTableName, cfg)
			if err != nil {
				return microerror.Mask(err)
			}
			rt := HostPostRouteTablesAdapterRoute{
				RouteTableName: routeTableName,
				RouteTableID:   routeTableID,
				// Requester CIDR block, we create the peering connection from the
				// guest's CIDR for being able to access Vault's ELB.
				CidrBlock:        key.ClusterNetworkCIDR(cfg.CustomObject),
				PeerConnectionID: peerConnectionID,
			}
			i.PublicRoutes = append(i.PublicRoutes, rt)
		}
	}
	return nil
}

type HostPostRouteTablesAdapterRoute struct {
	RouteTableName   string
	RouteTableID     string
	CidrBlock        string
	PeerConnectionID string
}

// waitForPeeringConnectionID keeps asking for the peering connection ID until it is obtained or
// a timeout expires. It is needed because the peering connection is created as part of the
// guest stack, we need to wait until the required resources from the guest are in place.
func waitForPeeringConnectionID(cfg Config) (string, error) {
	clusterID := key.ClusterID(cfg.CustomObject)
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("status-code"),
				Values: []*string{
					aws.String("active"),
				},
			},
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(clusterID),
				},
			},
		},
	}
	var peeringID string
	// TODO the logger should be injected and not magically made up in the depths
	// of the code where nobody knows about it and nobody has control over it.
	c := micrologger.Config{
		Caller:             micrologger.DefaultCaller,
		IOWriter:           micrologger.DefaultIOWriter,
		TimestampFormatter: micrologger.DefaultTimestampFormatter,
	}
	logger, err := micrologger.New(c)
	if err != nil {
		return "", microerror.Mask(err)
	}

	o := func() error {
		output, err := cfg.Clients.EC2.DescribeVpcPeeringConnections(input)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(output.VpcPeeringConnections) > 1 {
			return microerror.Maskf(tooManyResultsError, "peering connections")
		} else if len(output.VpcPeeringConnections) == 0 {
			return microerror.Maskf(notFoundError, "peering connections")
		}
		peeringID = *output.VpcPeeringConnections[0].VpcPeeringConnectionId
		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 10*time.Second)
	n := backoff.NewNotifier(logger, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return peeringID, nil
}

func routeTableID(name string, cfg Config) (string, error) {
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
	output, err := cfg.HostClients.EC2.DescribeRouteTables(input)
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
