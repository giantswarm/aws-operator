package cproutetables

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "cproutetables"
)

type Config struct {
	Logger       micrologger.Logger
	Installation string

	Names []string
}

type Resource struct {
	logger       micrologger.Logger
	installation string

	mutex       sync.Mutex
	routeTables []*ec2.RouteTable

	names []string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}

	// TODO: This is temporary fix until we get rid of routeTables flag.
	if len(config.Names) == 1 && config.Names[0] == "" {
		config.Names = []string{}
	}

	r := &Resource{
		logger:       config.Logger,
		installation: config.Installation,

		routeTables: []*ec2.RouteTable{},
		mutex:       sync.Mutex{},

		names: config.Names,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addRouteTablesToContext(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding cached route tables")
	if len(r.routeTables) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "found cached route tables")
		cc.Status.ControlPlane.RouteTables = r.routeTables

		return nil
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cached route tables")

	r.logger.LogCtx(ctx, "level", "debug", "message", "caching route tables")
	if len(r.names) == 0 {
		// We do not have the cached route tables, so we look them up using tags.
		tables, err := r.lookupByTag(ctx, cc.Client.ControlPlane.AWS.EC2, r.installation)
		if err != nil {
			return microerror.Mask(err)
		}

		r.routeTables = tables
	} else {
		// We do not have the cached route tables, so we look them up using names
		// supplied via RouteTables flag.
		for _, name := range r.names {
			rt, err := r.lookupByName(ctx, cc.Client.ControlPlane.AWS.EC2, name)
			if err != nil {
				return microerror.Mask(err)
			}

			r.routeTables = append(r.routeTables, rt)
		}
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "cached route tables")

	cc.Status.ControlPlane.RouteTables = r.routeTables

	return nil
}

func (r *Resource) lookupByName(ctx context.Context, client EC2, name string) (*ec2.RouteTable, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding route table ID for %#q", name))

	i := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}

	o, err := client.DescribeRouteTables(i)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(o.RouteTables) != 1 {
		return nil, microerror.Maskf(executionFailedError, "expected one route table, got %d", len(o.RouteTables))
	}

	rt := o.RouteTables[0]

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found route table for %#q", name))

	return rt, nil
}

func (r *Resource) lookupByTag(ctx context.Context, client EC2, installation string) ([]*ec2.RouteTable, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding route tables for installation %#q", installation))

	var privateTables []*ec2.RouteTable
	{
		i := &ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(installation),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagRouteTableType)),
					Values: []*string{
						aws.String("private"),
					},
				},
			},
		}

		o, err := client.DescribeRouteTables(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		privateTables = o.RouteTables
	}

	output := []*ec2.RouteTable{}
	{
		// filter out TCNP tables
		for _, table := range privateTables {
			isTCNPTable := false
			for _, tag := range table.Tags {
				if *tag.Key == key.TagStack && *tag.Value == "tcnp" {
					isTCNPTable = true
					break
				}
			}
			if !isTCNPTable {
				output = append(output, table)
			}
		}
	}

	if len(output) == 0 {
		return nil, microerror.Maskf(executionFailedError, "expected at least one route table, got 0")
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d route tables for installation %#q", len(output), installation))

	return output, nil
}
