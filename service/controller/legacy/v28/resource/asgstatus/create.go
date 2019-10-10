package asgstatus

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v28/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v28/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
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
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(fmt.Sprintf("%s-worker", key.ClusterID(cr))),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(key.ClusterID(cr)),
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
			r.logger.LogCtx(ctx, "level", "debug", "message", "worker asg not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "worker asg not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
		if len(o.Reservations[0].Instances) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "worker asg not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		asgName = awstags.ValueForKey(o.Reservations[0].Instances[0].Tags, "aws:autoscaling:groupName")
	}

	var asg *autoscaling.Group
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding ASG %#q", asgName))

		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				&asgName,
			},
		}

		o, err := cc.Client.TenantCluster.AWS.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.AutoScalingGroups) != 1 {
			return microerror.Maskf(executionFailedError, "there must be one item for ASG %#q", asgName)
		}
		asg = o.AutoScalingGroups[0]

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found ASG %#q", asgName))
	}

	var desiredCapacity int
	{
		if asg.DesiredCapacity == nil {
			return microerror.Maskf(executionFailedError, "desired capacity must not be empty for ASG %#q", asgName)
		}
		desiredCapacity = int(*asg.DesiredCapacity)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("desired capacity of %#q is %d", asgName, desiredCapacity))
	}

	var maxSize int
	{
		if asg.MaxSize == nil {
			return microerror.Maskf(executionFailedError, "max size must not be empty for ASG %#q", asgName)
		}
		maxSize = int(*asg.MaxSize)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("max size of %#q is %d", asgName, maxSize))
	}

	var minSize int
	{
		if asg.MinSize == nil {
			return microerror.Maskf(executionFailedError, "min size must not be empty for ASG %#q", asgName)
		}
		minSize = int(*asg.MinSize)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("min size of %#q is %d", asgName, minSize))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status with desired capacity")

		newObj, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(cr.GetNamespace()).Get(cr.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		if newObj.Status.Cluster.Scaling.DesiredCapacity != desiredCapacity {
			newObj.Status.Cluster.Scaling.DesiredCapacity = desiredCapacity
			_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(newObj.GetNamespace()).UpdateStatus(newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "updated status with desired capacity")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)

			return nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not update status with desired capacity")
		}
	}

	{
		cc.Status.TenantCluster.TCCP.ASG.DesiredCapacity = desiredCapacity
		cc.Status.TenantCluster.TCCP.ASG.MaxSize = maxSize
		cc.Status.TenantCluster.TCCP.ASG.MinSize = minSize
		cc.Status.TenantCluster.TCCP.ASG.Name = asgName
	}

	return nil
}
