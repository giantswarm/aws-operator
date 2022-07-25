package tcnpsecuritygroups

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
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

func (r *Resource) addInfoToCtx(ctx context.Context, cr infrastructurev1alpha3.AWSMachineDeployment) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var desired []string
	{
		r.logger.Debugf(ctx, "finding desired node pool security groups for machine deployment %#q", key.MachineDeploymentID(&cr))

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

		// Check if the Node Pool is being deleted. Therefore we check for the EC2
		// instances' state and ignore all security groups of Node Pools that do not
		// have any instance running.
		for _, sg := range o.SecurityGroups {
			var machineDeploymentID string
			{
				for _, tag := range sg.Tags {
					if *tag.Key == key.TagMachineDeployment {
						machineDeploymentID = *tag.Value
						break
					}
				}
			}

			// Ignore the security group of this Node Pool because it does not need an
			// ingress rule for itself.
			if machineDeploymentID == key.MachineDeploymentID(&cr) {
				continue
			}

			i := &ec2.DescribeInstancesInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String(fmt.Sprintf("tag:%s", key.TagMachineDeployment)),
						Values: []*string{
							aws.String(machineDeploymentID),
						},
					},
					{
						Name: aws.String("instance-state-name"),
						Values: []*string{
							aws.String(ec2.InstanceStateNamePending),
							aws.String(ec2.InstanceStateNameRunning),
							aws.String(ec2.InstanceStateNameStopped),
							aws.String(ec2.InstanceStateNameStopping),
						},
					},
				},
			}

			o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
			if err != nil {
				return microerror.Mask(err)
			}

			// Add the security group ID to the list only if the Node Pool has any
			// running instances. If there are no running or pending instances, the
			// Node Pool might be deleted and we want to remove the security group
			// rules as well.
			if len(o.Reservations) > 0 && len(o.Reservations[0].Instances) > 0 {
				desired = append(desired, *sg.GroupId)
			}
		}

		r.logger.Debugf(ctx, "found %d desired node pool security groups for machine deployment %#q", len(desired), key.MachineDeploymentID(&cr))
	}

	{
		cc.Spec.TenantCluster.TCNP.SecurityGroupIDs = desired
	}

	var current []string
	{
		var sg *ec2.SecurityGroup
		r.logger.Debugf(ctx, "finding current node pool security groups for machine deployment %#q", key.MachineDeploymentID(&cr))

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
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		if len(o.SecurityGroups) < 1 {
			r.logger.Debugf(ctx, "did not find current node pool security group for machine deployment %#q yet", key.MachineDeploymentID(&cr))
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		sg = o.SecurityGroups[0]

		// Iterate over all security group ingress rules in this Node Pool. We are
		// only interested in ingress rules that reference the security group of
		// this Node Pool. Note that rule.UserIdGroupPairs is empty for IP CIDR
		// based rules.
		for _, rule := range sg.IpPermissions {
			for _, gp := range rule.UserIdGroupPairs {
				// We are only interested in the "Node Pool to Node Pool" ingress rules.
				// The rule description is used for identifying the ingress rule. Thus
				// it must not change. Otherwise the tcnpsecuritygroups resource will
				// not be able to properly find the current and desired state of the
				// ingress rules.
				if gp.Description != nil && strings.Contains(*gp.Description, "Allow traffic from other Node Pool Security Groups to the Security Group of this Node Pool.") {
					current = append(current, *gp.GroupId)
				}
			}
		}

		r.logger.Debugf(ctx, "found %d current node pool security groups for machine deployment %#q", len(current), key.MachineDeploymentID(&cr))
	}

	{
		cc.Status.TenantCluster.TCNP.SecurityGroupIDs = current
	}

	return nil
}
