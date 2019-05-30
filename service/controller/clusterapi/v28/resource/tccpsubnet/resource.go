package tccpsubnet

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
)

const (
	Name = "tccpsubnetv27"
)

type Config struct {
	Logger micrologger.Logger
}

type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addSubnetsToContext(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// The tenant cluster VPC is a requirement to find its associated subnets and
	// route tables. In case the VPC ID is not yet tracked in the controller
	// context we return an error and cause the resource to be canceled.
	if cc.Status.TenantCluster.TCCP.VPC.ID == "" {
		return microerror.Mask(vpcNotFoundError)
	}

	{
		i := &ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(cc.Status.TenantCluster.TCCP.VPC.ID),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeRouteTables(i)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.TCCP.RouteTables = o.RouteTables
	}

	{
		i := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(cc.Status.TenantCluster.TCCP.VPC.ID),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSubnets(i)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.TCCP.Subnets = o.Subnets
	}

	return nil
}
