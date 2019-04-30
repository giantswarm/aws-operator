package routetable

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
)

const (
	Name = "routetablev27"
)

type Config struct {
	Logger micrologger.Logger

	Names []string
}

type Resource struct {
	logger micrologger.Logger

	// mappings is a mapping of route table names and IDs, where the key is the
	// name and the value is the ID.
	mappings map[string]string
	mutex    sync.Mutex

	names []string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Names == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Names must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		mappings: map[string]string{},
		mutex:    sync.Mutex{},

		names: config.Names,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addRouteTableMappingsToContext(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// We check if we have all mappings cached for the configured route table
	// names. If we find all information, we return them.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cached route table mappings")

		if len(r.mappings) == len(r.names) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found cached route table mappings")
			cc.Status.ControlPlane.RouteTable.Mappings = r.mappings

			return nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cached route table mappings")
		}
	}

	// We do not have the cached mappings, so we look them up.
	mappings := map[string]string{}
	for _, name := range r.names {
		id, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.EC2, name)
		if err != nil {
			return microerror.Mask(err)
		}

		mappings[name] = id
	}

	// At this point we found all route table mappings and can cache them
	// internally.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "caching route table mappings")
		r.mappings = mappings
		r.logger.LogCtx(ctx, "level", "debug", "message", "cached route table mappings")
	}

	cc.Status.ControlPlane.RouteTable.Mappings = mappings

	return nil
}

func (r *Resource) lookup(ctx context.Context, client EC2, name string) (string, error) {
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
		return "", microerror.Mask(err)
	}
	if len(o.RouteTables) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected one route table, got %d", len(o.RouteTables))
	}

	id := *o.RouteTables[0].RouteTableId

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found route table ID %#q for %#q", id, name))

	return id, nil
}
