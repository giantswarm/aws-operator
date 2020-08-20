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
	Logger micrologger.Logger

	Names []string
}

type Resource struct {
	logger micrologger.Logger

	mutex       sync.Mutex
	routeTables []*ec2.RouteTable

	names []string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if len(config.Names) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.Names must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

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

	// We check if we have all route tables cached for the configured route table
	// names. If we find all information, we return them.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cached route tables")

		if len(r.routeTables) == len(r.names) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found cached route tables")
			cc.Status.ControlPlane.RouteTables = r.routeTables

			return nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cached route tables")
		}
	}

	// We do not have the cached route tables, so we look them up.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "caching route tables")

		for _, name := range r.names {
			rt, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.EC2, name)
			if err != nil {
				return microerror.Mask(err)
			}

			r.routeTables = append(r.routeTables, rt)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "cached route tables")
	}

	cc.Status.ControlPlane.RouteTables = r.routeTables

	return nil
}

func (r *Resource) lookup(ctx context.Context, client EC2, name string) (*ec2.RouteTable, error) {
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

	if len(o.RouteTables) == 0 {
		return nil, microerror.Maskf(executionFailedError, "expected at least one route table, got %d", len(o.RouteTables))
	}

	rt := o.RouteTables[0]

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found route table for %#q", name))

	return rt, nil
}
