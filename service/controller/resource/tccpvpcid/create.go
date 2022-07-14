package tccpvpcid

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v12/service/controller/key"
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

	var vpcs []*ec2.Vpc
	{
		r.logger.Debugf(ctx, "finding vpcs")

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

		r.logger.Debugf(ctx, "found vpcs")
	}

	{
		r.logger.Debugf(ctx, "finding vpc id for tenant cluster %#q", key.ClusterID(&cr))

		if len(vpcs) > 1 {
			return microerror.Maskf(executionFailedError, "expected one vpc, got %d", len(vpcs))
		}

		if len(vpcs) < 1 {
			r.logger.Debugf(ctx, "did not find vpc id for tenant cluster %#q", key.ClusterID(&cr))
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		r.logger.Debugf(ctx, "found vpc id %#q for tenant cluster %#q", *vpcs[0].VpcId, key.ClusterID(&cr))

		cc.Status.TenantCluster.TCCP.VPC.ID = *vpcs[0].VpcId
	}

	return nil
}
