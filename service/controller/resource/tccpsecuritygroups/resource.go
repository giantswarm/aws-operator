package tccpsecuritygroups

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "tccpsecuritygroups"
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

		if len(groups) > 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(groups))
		}

		if len(groups) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find security groups for tenant cluster %#q yet", key.ClusterID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found security groups for tenant cluster %#q", key.ClusterID(&cr)))
	}

	{
		cc.Status.TenantCluster.TCCP.SecurityGroups = groups
	}

	return nil
}
