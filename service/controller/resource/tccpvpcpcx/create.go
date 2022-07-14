package tccpvpcpcx

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

	{
		if cc.Status.TenantCluster.TCCP.VPC.ID == "" {
			r.logger.Debugf(ctx, "tenant cluster vpc id not available yet")
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}
	}

	var vpcPCXs []*ec2.VpcPeeringConnection
	{
		r.logger.Debugf(ctx, "finding vpc peering connections")

		i := &ec2.DescribeVpcPeeringConnectionsInput{
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
				{
					Name: aws.String("requester-vpc-info.vpc-id"),
					Values: []*string{
						aws.String(cc.Status.TenantCluster.TCCP.VPC.ID),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeVpcPeeringConnections(i)
		if err != nil {
			return microerror.Mask(err)
		}

		vpcPCXs = o.VpcPeeringConnections

		r.logger.Debugf(ctx, "found vpc peering connections")
	}

	{
		r.logger.Debugf(ctx, "finding vpc peering connection id for tenant cluster %#q", key.ClusterID(&cr))

		if len(vpcPCXs) > 1 {
			return microerror.Maskf(executionFailedError, "expected one vpc peering connection, got %d", len(vpcPCXs))
		}

		if len(vpcPCXs) < 1 {
			r.logger.Debugf(ctx, "did not find vpc peering connection id for tenant cluster %#q", key.ClusterID(&cr))
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		r.logger.Debugf(ctx, "found vpc peering connection id %#q for tenant cluster %#q", *vpcPCXs[0].VpcPeeringConnectionId, key.ClusterID(&cr))

		cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID = *vpcPCXs[0].VpcPeeringConnectionId
	}

	return nil
}
