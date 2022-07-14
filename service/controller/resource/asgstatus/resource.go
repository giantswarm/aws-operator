package asgstatus

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v2/service/controller/controllercontext"
)

const (
	Name = "asgstatus"
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
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var asgName string
	{
		if cc.Status.TenantCluster.ASG.Name == "" {
			r.logger.Debugf(ctx, "auto scaling group name not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		asgName = cc.Status.TenantCluster.ASG.Name
	}

	var asg *autoscaling.Group
	{
		r.logger.Debugf(ctx, "finding auto scaling group %#q", asgName)

		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String(asgName),
			},
		}

		o, err := cc.Client.TenantCluster.AWS.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.AutoScalingGroups) != 1 {
			return microerror.Maskf(executionFailedError, "there must be one item for auto scaling group %#q", asgName)
		}
		asg = o.AutoScalingGroups[0]

		r.logger.Debugf(ctx, "found auto scaling group %#q", asgName)
	}

	var desiredCapacity int
	{
		if asg.DesiredCapacity == nil {
			return microerror.Maskf(executionFailedError, "desired capacity must not be empty for auto scaling group %#q", asgName)
		}
		desiredCapacity = int(*asg.DesiredCapacity)
		r.logger.Debugf(ctx, "desired capacity of %#q is %d", asgName, desiredCapacity)
	}

	var maxSize int
	{
		if asg.MaxSize == nil {
			return microerror.Maskf(executionFailedError, "max size must not be empty for auto scaling group %#q", asgName)
		}
		maxSize = int(*asg.MaxSize)
		r.logger.Debugf(ctx, "max size of %#q is %d", asgName, maxSize)
	}

	var minSize int
	{
		if asg.MinSize == nil {
			return microerror.Maskf(executionFailedError, "min size must not be empty for auto scaling group %#q", asgName)
		}
		minSize = int(*asg.MinSize)
		r.logger.Debugf(ctx, "min size of %#q is %d", asgName, minSize)
	}

	{
		cc.Status.TenantCluster.ASG.DesiredCapacity = desiredCapacity
		cc.Status.TenantCluster.ASG.MaxSize = maxSize
		cc.Status.TenantCluster.ASG.MinSize = minSize
	}

	return nil
}
