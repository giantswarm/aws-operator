package asg

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/asg/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface

	Stack        string
	TagKey       string
	TagValueFunc func(cr key.LabelsGetter) string
}

type ASG struct {
	k8sClient k8sclient.Interface

	asgsCache      *cache.ASGs
	instancesCache *cache.Instances

	stack        string
	tagKey       string
	tagValueFunc func(cr key.LabelsGetter) string
}

func New(config Config) (*ASG, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
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

	a := &ASG{
		k8sClient: config.K8sClient,

		asgsCache:      cache.NewASGs(),
		instancesCache: cache.NewInstances(),

		stack:        config.Stack,
		tagKey:       config.TagKey,
		tagValueFunc: config.TagValueFunc,
	}

	return a, nil
}

func (a *ASG) Drainable(ctx context.Context, obj interface{}) (string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// In order to fetch all the relevant ASGs we have to fetch all relevant
	// instances. The Autoscaling API does not provide tag filters but only
	// filtering by ASG names. Therefore we get the ASG names from the instances,
	// because the EC2 API has tag filter support.
	instances, err := a.cachedInstances(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Having the EC2 instances we can parse the ASG names from the instance tags
	// and fetch the ASGs in order to get the lifecycle hook information we are
	// interested it.
	var asgs []*autoscaling.Group
	{
		var names []string
		{
			m := map[string]struct{}{}

			for _, i := range instances {
				m[asgNameFromInstance(i)] = struct{}{}
			}

			for k := range m {
				names = append(names, k)
			}
		}

		asgs, err = a.cachedASGs(ctx, cr, names)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	// The user asks for the next drainable ASG. There is 1 ASG only, for instance
	// for Node Pools and Single Master setups. There are 3 ASGs in HA Masters
	// setups. So in case there are multiple ASGs we want to be careful and only
	// drain instances of one of them in case in each of the other ASGs we have at
	// least one healthy instance. This aims to ensure that an HA Masters setup is
	// most reliable.
	if len(asgs) > 1 {
		var c int

		for _, a := range asgs {
			for _, i := range a.Instances {
				if *i.LifecycleState == autoscaling.LifecycleStateInService {
					c++
					break
				}
			}
		}

		if c < len(asgs) {
			return "", microerror.Mask(notFoundError)
		}
	}

	// At this point we return the first drainable ASG, which we identify based on
	// the Terminating:Wait condition of its instances. This should be generic
	// enough to cover all of our cases for Node Pools and Control Planes having 1
	// to N EC2 instances configured.
	for _, a := range asgs {
		for _, i := range a.Instances {
			if *i.LifecycleState == autoscaling.LifecycleStateTerminatingWait {
				return *a.AutoScalingGroupName, nil
			}
		}
	}

	return "", microerror.Mask(notFoundError)
}

func (a *ASG) cachedASGs(ctx context.Context, cr metav1.Object, names []string) ([]*autoscaling.Group, error) {
	var err error
	var ok bool

	var asgs []*autoscaling.Group
	{
		ck := a.asgsCache.Key(ctx, cr)

		if ck == "" {
			asgs, err = a.lookupASGs(ctx, names)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			asgs, ok = a.asgsCache.Get(ctx, ck)
			if !ok {
				asgs, err = a.lookupASGs(ctx, names)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				a.asgsCache.Set(ctx, ck, asgs)
			}
		}
	}

	return asgs, nil
}

func (a *ASG) cachedInstances(ctx context.Context, cr metav1.Object) ([]*ec2.Instance, error) {
	var err error
	var ok bool

	var instances []*ec2.Instance
	{
		ck := a.instancesCache.Key(ctx, cr)

		if ck == "" {
			instances, err = a.lookupInstances(ctx, cr)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			instances, ok = a.instancesCache.Get(ctx, ck)
			if !ok {
				instances, err = a.lookupInstances(ctx, cr)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				a.instancesCache.Set(ctx, ck, instances)
			}
		}
	}

	return instances, nil
}

func (a *ASG) lookupASGs(ctx context.Context, names []string) ([]*autoscaling.Group, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var asgs []*autoscaling.Group
	{
		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: toPtrList(names),
		}

		o, err := cc.Client.TenantCluster.AWS.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		asgs = o.AutoScalingGroups
	}

	return asgs, nil
}

func (a *ASG) lookupInstances(ctx context.Context, cr metav1.Object) ([]*ec2.Instance, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

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
					aws.String(a.stack),
				},
			},
			{
				Name: aws.String(fmt.Sprintf("tag:%s", a.tagKey)),
				Values: []*string{
					aws.String(a.tagValueFunc(cr)),
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
		return nil, microerror.Mask(err)
	}

	var instances []*ec2.Instance
	for _, r := range o.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, i)
		}
	}

	if len(instances) == 0 {
		return nil, microerror.Mask(notFoundError)
	}

	return instances, nil
}

func asgNameFromInstance(i *ec2.Instance) string {
	return awstags.ValueForKey(i.Tags, "aws:autoscaling:groupName")
}

func toPtrList(l []string) []*string {
	var p []*string

	for _, s := range l {
		p = append(p, to.StringP(s))
	}

	return p
}
