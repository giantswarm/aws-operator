package tcnpsecuritygroups

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "tcnpsecuritygroups"
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

func (r *Resource) addInfoToCtx(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var groups []*ec2.SecurityGroup
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding all np security groups for tenant cluster %#q", key.ClusterID(&cr)))

		i := &ec2.DescribeSecurityGroupsInput{
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
						aws.String(key.StackTCNP),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		groups = o.SecurityGroups

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found np security groups for tenant cluster %#q", key.ClusterID(&cr)))
	}

	{
		cc.Spec.TenantCluster.TCNP.SecurityGroups = groups
	}

	var securityGroupIDs []string
	{
		// list single security group of this very node pool
		// get all ingress rules from this security group
		// put all security groups referenced in the ingress rules into controller context status
		var sg *ec2.SecurityGroup
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding general np security group for tenant cluster %#q", key.ClusterID(&cr)))

		// use tag filter from previous security groups list
		i := &ec2.DescribeSecurityGroupsInput{
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
						aws.String(key.StackTCNP),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagMachineDeployment)),
					Values: []*string{
						aws.String(key.MachineDeploymentID(&cr)),
					},
				},
			},
		}
		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.SecurityGroups) > 1 {
			return microerror.Maskf(executionFailedError, "expected one security groups, got %d", len(o.SecurityGroups))
		}

		if len(o.SecurityGroups) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find  worker security group for machine deployment %#q yet", key.MachineDeploymentID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		sg = o.SecurityGroups[0]

		// iterate over all security groups ingress rules
		for _, sgRule := range sg.IpPermissions {
			// we are only interested in ingress rules that uses security groups IDs reference
			// sgRule.UserIdGroupPairs is empty for IP CIDR based rules
			for _, gp := range sgRule.UserIdGroupPairs {
				securityGroupIDs = append(securityGroupIDs, *gp.GroupId)
			}
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finded %d general node pool worker security groups ingress rules for machine deployment %#q", len(securityGroupIDs), key.MachineDeploymentID(&cr)))
	}

	{
		cc.Status.TenantCluster.TCNP.SecurityGroupIDs = securityGroupIDs
	}

	return nil
}
