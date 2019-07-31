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
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var groups []*ec2.SecurityGroup
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding security groups for tenant cluster %#q", key.ClusterID(&cr)))

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.SecurityGroupName(&cr, "ingress")),
					},
				},
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.SecurityGroupName(&cr, "master")),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		groups = o.SecurityGroups

		if len(groups) > 2 {
			return microerror.Maskf(executionFailedError, "expected two security groups, got %d", len(groups))
		}

		if len(groups) < 2 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find security groups for tenant cluster %#q yet", key.ClusterID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found security groups for tenant cluster %#q", key.ClusterID(&cr)))
	}

	{
		cc.Status.TenantCluster.TCCP.SecurityGroup.Ingress.ID = idFromGroups(groups, key.SecurityGroupName(&cr, "ingress"))
		cc.Status.TenantCluster.TCCP.SecurityGroup.Master.ID = idFromGroups(groups, key.SecurityGroupName(&cr, "master"))
	}

	return nil
}

func idFromGroups(groups []*ec2.SecurityGroup, name string) string {
	for _, g := range groups {
		if valueForKey(g.Tags, "Name") == name {
			return *g.GroupId
		}
	}

	return ""
}

func valueForKey(tags []*ec2.Tag, key string) string {
	for _, t := range tags {
		if *t.Key == key {
			return *t.Value
		}
	}

	return ""
}
