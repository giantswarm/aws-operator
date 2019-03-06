package routetable

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	EC2    EC2
	Logger micrologger.Logger

	// Names are the route table names used to lookup IDs.
	Names []string
}

// RouteTable is a service implementation configured with the private route
// table names of the control plane. They never change during runtime. Before we
// fetched the route table IDs for the CPF stack management over and over again.
// This implementation here is to fetch them once and have them generally
// available.
type RouteTable struct {
	ec2    EC2
	logger micrologger.Logger

	// ids is a mapping of route table names and IDs, where the key is the name
	// and the value is the ID.
	ids   map[string]string
	mutex sync.Mutex

	names []string
}

// New creates a new route table service that has to be booted using Boot to
// cache the configured route table IDs associated with their names.
func New(config Config) (*RouteTable, error) {
	if config.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.EC2 must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &RouteTable{
		ec2:    config.EC2,
		logger: config.Logger,

		ids:   map[string]string{},
		mutex: sync.Mutex{},

		names: config.Names,
	}

	return r, nil
}

func (r *RouteTable) IDForName(ctx context.Context, name string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	id, ok := r.ids[name]
	if ok {
		return id, nil
	}

	id, err := r.searchID(ctx, name)
	if err != nil {
		return "", microerror.Mask(err)
	}
	r.ids[name] = id

	return id, nil
}

func (r *RouteTable) Names() []string {
	return r.names
}

func (r *RouteTable) searchID(ctx context.Context, name string) (string, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", "finding route table ID for %#q", name)

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
	o, err := r.ec2.DescribeRouteTables(i)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(o.RouteTables) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected one route table, got %d", len(o.RouteTables))
	}

	id := *o.RouteTables[0].RouteTableId

	r.logger.LogCtx(ctx, "level", "debug", "message", "found route table ID %#q for %#q", id, name)

	return id, nil
}
