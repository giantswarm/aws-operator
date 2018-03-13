package adapter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/awsconfig/v9/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v9/templates/cloudformation/hostpost/route_tables.go
//

type hostRouteTablesAdapter struct {
	RouteTables []RouteTable
}

type RouteTable struct {
	Name             string
	RouteTableID     string
	CidrBlock        string
	PeerConnectionID string
}

func (i *hostRouteTablesAdapter) getHostRouteTables(cfg Config) error {
	peerConnectionID, err := waitForPeeringConnectionID(cfg)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, routeTableName := range cfg.CustomObject.Spec.AWS.VPC.RouteTableNames {
		routeTableID, err := routeTableID(routeTableName, cfg)
		if err != nil {
			return microerror.Mask(err)
		}
		rt := RouteTable{
			Name:         routeTableName,
			RouteTableID: routeTableID,
			// Requester CIDR block, we create the peering connection from the guest's private subnet.
			CidrBlock:        key.PrivateSubnetCIDR(cfg.CustomObject),
			PeerConnectionID: peerConnectionID,
		}
		i.RouteTables = append(i.RouteTables, rt)
	}

	return nil
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
	queryOperation := func() error {
		output, err := cfg.Clients.EC2.DescribeVpcPeeringConnections(input)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(output.VpcPeeringConnections) > 1 {
			return microerror.Maskf(tooManyResultsError, "peering connections")
		}
		peeringID = *output.VpcPeeringConnections[0].VpcPeeringConnectionId
		return nil
	}
	queryNotify := func(err error, delay time.Duration) {
		logger.Log("error", fmt.Sprintf("query VPC peering connection ID failed, retrying with delay %.0fm%.0fs: '%#v'", delay.Minutes(), delay.Seconds(), err))
	}
	bo := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      2 * time.Minute,
		Clock:               backoff.SystemClock,
	}
	if err := backoff.RetryNotify(queryOperation, bo, queryNotify); err != nil {
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
	if len(output.RouteTables) > 1 {
		return "", microerror.Maskf(tooManyResultsError, "route tables: %s", name)
	}

	return *output.RouteTables[0].RouteTableId, nil
}
