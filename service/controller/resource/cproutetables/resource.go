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
	// expectedTableCount is set depending on the collection method. If
	// filtering by tag yields result, it defaults to the count of RouteTables
	// returned this way. Otherwise it is set to len(names).
	expectedTableCount int

	names []string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// TODO: Get installation from flags
	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}

	r := &Resource{
		logger:       config.Logger,
		installation: config.Installation,

		routeTables:        []*ec2.RouteTable{},
		mutex:              sync.Mutex{},
		expectedTableCount: 0,

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

	if len(r.routeTables) == r.expectedTableCount && r.expectedTableCount > 0 {
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
	r.expectedTableCount = len(r.routeTables)

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

	i := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:giantswarm.io/cluster"),
				Values: []*string{
					aws.String(installation),
				},
			},
			{
				Name: aws.String("tag:giantswarm.io/route-table-type"),
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
	if len(o.RouteTables) == 0 {
		return nil, microerror.Maskf(executionFailedError, "expected at least one route table, got 0")
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d route tables for installation %#q", len(o.RouteTables), installation))

	return o.RouteTables, nil
}
