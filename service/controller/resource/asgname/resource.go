package asgname

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/v13/pkg/awstags"
	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

const (
	Name = "asgname"
)

type Config struct {
	Logger micrologger.Logger

	Stack        string
	TagKey       string
	TagValueFunc func(cr key.LabelsGetter) string
}

type Resource struct {
	logger micrologger.Logger

	stack        string
	tagKey       string
	tagValueFunc func(cr key.LabelsGetter) string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Stack == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Stack must not be empty", config)
	}
	if config.TagKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TagKey must not be empty", config)
	}
	if config.TagValueFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TagValueFunc must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		stack:        config.Stack,
		tagKey:       config.TagKey,
		tagValueFunc: config.TagValueFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.Debugf(ctx, "finding auto scaling group name")

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
						aws.String(r.stack),
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
			r.logger.Debugf(ctx, "auto scaling group not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			r.logger.Debugf(ctx, "auto scaling group not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}
		if len(o.Reservations[0].Instances) == 0 {
			r.logger.Debugf(ctx, "auto scaling group not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		cc.Status.TenantCluster.ASG.Name = awstags.ValueForKey(o.Reservations[0].Instances[0].Tags, "aws:autoscaling:groupName")

		r.logger.Debugf(ctx, "found auto scaling group name %#q", cc.Status.TenantCluster.ASG.Name)
	}

	return nil
}
