package tccpvpcid

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
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

	var vpcs []*ec2.Vpc
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding vpcs")

		i := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
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
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeVpcs(i)
		if err != nil {
			return microerror.Mask(err)
		}

		vpcs = o.Vpcs

		r.logger.LogCtx(ctx, "level", "debug", "message", "found vpcs")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding vpc id for tenant cluster %#q", key.ClusterID(&cr)))

		if len(vpcs) > 1 {
			return microerror.Maskf(executionFailedError, "expected one vpc, got %d", len(vpcs))
		}

		if len(vpcs) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find vpc id for tenant cluster %#q", key.ClusterID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found vpc id %#q for tenant cluster %#q", *vpcs[0].VpcId, key.ClusterID(&cr)))

		cc.Status.TenantCluster.TCCP.VPC.ID = *vpcs[0].VpcId
	}

	return nil
}
