package tcnpinstanceinfo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v12/service/controller/key"
)

const (
	Name = "tcnpinstanceinfo"
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

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	i := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
				Values: []*string{
					aws.String(key.ClusterID(&cr)),
				},
			},
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagMachineDeployment)),
				Values: []*string{
					aws.String(key.MachineDeploymentID(&cr)),
				},
			},
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagStack)),
				Values: []*string{
					aws.String(key.StackTCNP),
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
	if IsNotFound(err) {
		r.logger.Debugf(ctx, "worker asg not available yet")
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	if len(o.Reservations) == 0 {
		r.logger.Debugf(ctx, "worker asg not available yet")
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}
	if len(o.Reservations[0].Instances) == 0 {
		r.logger.Debugf(ctx, "worker asg not available yet")
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	numberOfSpotInstances := 0
	instanceTypes := []string{}

	for _, reservation := range o.Reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceLifecycle != nil && *instance.InstanceLifecycle == "spot" {
				numberOfSpotInstances++
			}
			if instance.InstanceType != nil && !containsString(instanceTypes, *instance.InstanceType) {
				instanceTypes = append(instanceTypes, *instance.InstanceType)
			}
		}
	}

	{
		cc.Status.TenantCluster.TCNP.Instances.InstanceTypes = instanceTypes
		cc.Status.TenantCluster.TCNP.Instances.NumberOfSpotInstances = numberOfSpotInstances
	}

	return nil
}

func containsString(list []string, match string) bool {
	for _, s := range list {
		if s == match {
			return true
		}
	}

	return false
}
