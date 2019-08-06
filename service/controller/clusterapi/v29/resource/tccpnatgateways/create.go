package tccpnatgateways

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not yet availabile")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

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
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding natgateways")

		i := &ec2.DescribeNatGatewaysInput{
			Filter: []*ec2.Filter{
				{
					Name: aws.String(key.TagCluster),
					Values: []*string{
						aws.String(key.ClusterID(&cr)),
					},
				},
				{
					Name: aws.String(key.TagStack),
					Values: []*string{
						aws.String(key.StackTCCP),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeNatGateways(i)
		if err != nil {
			return microerror.Mask(err)
		}

		natgateways = o.NatGateways

		r.logger.LogCtx(ctx, "level", "debug", "message", "found natgateways")
	}

	{
		if len(natgateways) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find natgateways for tenant cluster %#q", key.ClusterID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d natgateways for tenant cluster %#q", len(natgateways), key.ClusterID(&cr)))

		cc.Status.TenantCluster.TCCP.NATGateways = natgateways
	}

	return nil
}
