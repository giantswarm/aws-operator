package tccpsecuritygroup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

const (
	Name = "tccpsecuritygroupv29"
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

func (r *Resource) addInfoToCtx(ctx context.Context, cr v1alpha1.MachineDeployment) error {
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
		cc.Status.TenantCluster.TCCP.SecurityGroup.Master.Permissions = permissionsFromGroups(groups, key.SecurityGroupName(&cr, "master"))
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

func permissionsFromGroups(groups []*ec2.SecurityGroup, name string) []*ec2.IpPermission {
	for _, g := range groups {
		if valueForKey(g.Tags, "Name") == name {
			return g.IpPermissions
		}
	}

	return nil
}

func valueForKey(tags []*ec2.Tag, key string) string {
	for _, t := range tags {
		if *t.Key == key {
			return *t.Value
		}
	}

	return ""
}
