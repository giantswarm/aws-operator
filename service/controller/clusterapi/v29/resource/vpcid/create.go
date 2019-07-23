package vpcid

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
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(&cr)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding vpc id for %s", clusterID))

	var vpcs []*ec2.Vpc
	{
		i := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(clusterID),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeVpcs(i)
		if err != nil {
			return microerror.Mask(err)
		}
		vpcs = o.Vpcs
	}

	{
		if len(vpcs) > 1 {
			return microerror.Maskf(tooManyResultsError, "expected one vpc, got %d", len(vpcs))
		}

		if len(vpcs) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find vpc id for %s", clusterID))
		}

		if len(vpcs) == 1 {
			if vpcs[0].VpcId != nil {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found vpc id %s for %s", *vpcs[0].VpcId, clusterID))

				cc.Status.TenantCluster.TCCP.VPC.ID = *vpcs[0].VpcId
			}
		}
	}

	return nil
}
