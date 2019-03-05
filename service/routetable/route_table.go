package routetable

import (
	"context"

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

type RouteTable struct {
	ec2    EC2
	logger micrologger.Logger

	// ids is a mapping of route table names and IDs, where the key is the name
	// and the value is the ID.
	ids map[string]string

	names []string
}

// New creates a new route table service that has to be booted using Boot to
// cache the confiured route table IDs associated with their names.
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

		ids: map[string]string{},

		names: config.Names,
	}

	return r, nil
}

func (r *RouteTable) Boot(ctx context.Context) error {
	for _, name := range r.names {
		id, err := r.searchID(name)
		if err != nil {
			return microerror.Mask(err)
		}

		r.ids[name] = id
	}

	return nil
}

func (r *RouteTable) IdForName(name string) (string, error) {
	id, ok := r.ids[name]
	if !ok {
		return "", microerror.Maskf(notFoundError, name)
	}

	return id, nil
}

func (r *RouteTable) searchID(name string) (string, error) {
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

	return *o.RouteTables[0].RouteTableId, nil
}
