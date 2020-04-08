package asgname

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "asgname"
)

type Config struct {
	Logger micrologger.Logger

	TagKey       string
	TagValueFunc func(cr key.LabelsGetter) string
}

type Resource struct {
	logger micrologger.Logger

	tagKey       string
	tagValueFunc func(cr key.LabelsGetter) string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TagKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TagKey must not be empty", config)
	}
	if config.TagValueFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TagValueFunc must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		tagKey:       config.TagKey,
		tagValueFunc: config.TagValueFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToLabelsGetter(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var asgName string
	{
		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(key.ClusterID(cr)),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagStack)),
					Values: []*string{
						aws.String(key.StackTCNP),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", r.tagKey)),
					Values: []*string{
						aws.String(r.tagValueFunc(cr)),
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
			r.logger.LogCtx(ctx, "level", "debug", "message", "asg not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "asg not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
		if len(o.Reservations[0].Instances) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "asg not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		asgName = awstags.ValueForKey(o.Reservations[0].Instances[0].Tags, "aws:autoscaling:groupName")
	}

	{
		cc.Status.TenantCluster.ASG.Name = asgName
	}

	return nil
}
