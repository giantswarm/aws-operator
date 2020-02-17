package tcnpsecuritygroups

import (
	"context"
	"fmt"
	"strings"

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

	var desiredSecurityGroupIDs []string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding desired node pool security groups for machine deployment %#q", key.MachineDeploymentID(&cr)))

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

		// check if the node pools is not being deleted
		// we check for ec2 instances for that node pool machine deployment
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
			// ignore security group for this node pool machine deployment
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

			// add security group to the list only if the node pool machine deployment has any running instances
			// if there are no running or pending instances the node pool machine deployment might be deleted
			// and we want to remove the sg rules as well
			if len(o.Reservations) > 0 && len(o.Reservations[0].Instances) > 0 {
				desiredSecurityGroupIDs = append(desiredSecurityGroupIDs, *sg.GroupId)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d desired node pool security groups for machine deployment %#q", len(desiredSecurityGroupIDs), key.MachineDeploymentID(&cr)))
	}

	{
		cc.Spec.TenantCluster.TCNP.SecurityGroupIDs = desiredSecurityGroupIDs
	}

	var currentSecurityGroupIDs []string
	{
		var sg *ec2.SecurityGroup
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding current node pool security groups for machine deployment %#q", key.MachineDeploymentID(&cr)))

		// TODO use tag filter from previous security desiredSecurityGroupIDs list
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
			return microerror.Maskf(executionFailedError, "expected one security desiredSecurityGroupIDs, got %d", len(o.SecurityGroups))
		}

		if len(o.SecurityGroups) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find current node pool security group for machine deployment %#q yet", key.MachineDeploymentID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		sg = o.SecurityGroups[0]

		// iterate over all security desiredSecurityGroupIDs ingress rules
		for _, sgRule := range sg.IpPermissions {
			// we are only interested in ingress rules that uses security desiredSecurityGroupIDs IDs reference
			// sgRule.UserIdGroupPairs is empty for IP CIDR based rules
			for _, gp := range sgRule.UserIdGroupPairs {
				// only NodePool to NodePool ingress rules are important
				if gp.Description != nil && strings.Contains(*gp.Description, "NodePoolToNodePool") {
					currentSecurityGroupIDs = append(currentSecurityGroupIDs, *gp.GroupId)
				}
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d current node pool security groups for machine deployment %#q", len(currentSecurityGroupIDs), key.MachineDeploymentID(&cr)))
	}

	{
		cc.Status.TenantCluster.TCNP.SecurityGroupIDs = currentSecurityGroupIDs
	}

	return nil
}
