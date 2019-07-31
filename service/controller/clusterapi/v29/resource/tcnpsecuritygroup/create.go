package tcnpsecuritygroup

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
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var id string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding security group ID for node pool %#q", key.MachineDeploymentID(&cr)))

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String("NodePoolSecurityGroup"),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.SecurityGroups) > 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		if len(o.SecurityGroups) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find security group id for node pool %#q yet", key.MachineDeploymentID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		id = *o.SecurityGroups[0].GroupId

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found security group id %#q for tenant cluster %#q", id, key.ClusterID(&cr)))
	}

	var ipPermissions []*ec2.IpPermission
	{
		p := &ec2.IpPermission{
			FromPort:   aws.Int64(-1),
			IpProtocol: aws.String("-1"),
			ToPort:     aws.Int64(-1),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					Description: aws.String("Allow traffic from the TCNP Node Pool Security Group to the TCCP Master Security Group."),
					GroupId:     aws.String(id),
				},
			},
		}

		ipPermissions = append(cc.Status.TenantCluster.TCCP.SecurityGroup.Master.Permissions, p)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating master security group for tenant cluster %#q", key.ClusterID(&cr)))

		i := &ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
			GroupId:       aws.String(cc.Status.TenantCluster.TCCP.SecurityGroup.Master.ID),
			IpPermissions: ipPermissions,
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.UpdateSecurityGroupRuleDescriptionsIngress(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated master security group for tenant cluster %#q", key.ClusterID(&cr)))
	}

	return nil
}
