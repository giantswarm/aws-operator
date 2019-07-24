package tccpsecuritygroupid

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
	md, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var groups []*ec2.SecurityGroup
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding ingress security groups")

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.SecurityGroupName(&md, "ingress")),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		groups = o.SecurityGroups

		r.logger.LogCtx(ctx, "level", "debug", "message", "found ingress security groups")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding ingress security group id for tenant cluster %#q", key.ClusterID(&md)))

		if len(groups) > 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(groups))
		}

		if len(groups) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find ingress security group for tenant cluster %#q", key.ClusterID(&md)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found ingress security group id %#q for tenant cluster %#q", *groups[0].GroupId, key.ClusterID(&md)))

		cc.Status.TenantCluster.TCCP.SecurityGroup.Ingress.ID = *groups[0].GroupId
	}

	return nil
}
