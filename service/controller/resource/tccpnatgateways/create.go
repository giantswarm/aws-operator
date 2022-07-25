package tccpnatgateways

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if IsNotFound(err) {
		r.logger.Debugf(ctx, "cluster cr not available yet")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var natgateways []*ec2.NatGateway
	{
		r.logger.Debugf(ctx, "finding natgateways")

		i := &ec2.DescribeNatGatewaysInput{
			Filter: []*ec2.Filter{
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(key.ClusterID(&cr)),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagStack)),
					Values: []*string{
						aws.String(key.StackTCCP),
					},
				},
				// ignore NAT gateway in state 'deleting' or 'deleted'
				{
					Name: aws.String("state"),
					Values: []*string{
						aws.String("available"),
						aws.String("pending"),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeNatGateways(i)
		if err != nil {
			return microerror.Mask(err)
		}

		natgateways = o.NatGateways

		r.logger.Debugf(ctx, "found natgateways")
	}

	{
		if len(natgateways) < 1 {
			r.logger.Debugf(ctx, "did not find natgateways for tenant cluster %#q", key.ClusterID(&cr))
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		r.logger.Debugf(ctx, "found %d natgateways for tenant cluster %#q", len(natgateways), key.ClusterID(&cr))

		cc.Status.TenantCluster.TCCP.NATGateways = natgateways
	}

	return nil
}
