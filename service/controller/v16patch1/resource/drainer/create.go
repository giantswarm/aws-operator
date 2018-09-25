package drainer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/v16patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1/key"
)

// EnsureCreated creates DrainerConfigs for ASG instances in terminating/wait
// state then lets node-operator to do its job.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	controllerCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	workerASGName := controllerCtx.Status.Drainer.WorkerASGName
	if workerASGName == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "worker ASG name is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var instances []*autoscaling.Instance
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding the guest cluster nodes being in state %#q", autoscaling.LifecycleStateTerminatingWait))

		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String(workerASGName),
			},
		}

		o, err := controllerCtx.AWSClient.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, g := range o.AutoScalingGroups {
			for _, i := range g.Instances {
				if *i.LifecycleState == autoscaling.LifecycleStateTerminatingWait {
					instances = append(instances, i)
				}
			}
		}

		if len(instances) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the guest cluster nodes being in state %#q", autoscaling.LifecycleStateTerminatingWait))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d guest cluster nodes being in state %#q", len(instances), autoscaling.LifecycleStateTerminatingWait))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring drainer configs for %d guest cluster nodes being in state %#q", len(instances), autoscaling.LifecycleStateTerminatingWait))

		for _, instance := range instances {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding drainer config for guest cluster nodes %#q", *instance.InstanceId))

			privateDNS, err := r.privateDNSForInstance(ctx, *instance.InstanceId)
			if err != nil {
				return microerror.Mask(err)
			}
			if privateDNS == "" {
				// It might happen that state transitioning within EC2 happen while we
				// try to gather information. An EC2 instance might be in
				// Terminating:Wait state and then moves to Terminated before we get a
				// chance to gather the drainer configs here. The operator then did its
				// job already and we only have to deal with the edge case situation. So
				// we just stop here and move on with the other instances.
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no private DNS for instance %#q", *instance.InstanceId))
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not draining instance %#q", *instance.InstanceId))
				continue
			}

			n := customObject.GetNamespace()
			o := metav1.GetOptions{}

			_, err = r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Get(privateDNS, o)
			if errors.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find drainer config for guest cluster node %#q", *instance.InstanceId))

				err := r.createDrainerConfig(ctx, customObject, *instance.InstanceId, privateDNS)
				if err != nil {
					return microerror.Mask(err)
				}

			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found drainer config for guest cluster node %#q", *instance.InstanceId))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured drainer configs for %d guest cluster nodes being in state %#q", len(instances), autoscaling.LifecycleStateTerminatingWait))
	}

	return nil
}

func (r *Resource) createDrainerConfig(ctx context.Context, customObject providerv1alpha1.AWSConfig, instanceID, privateDNS string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating drainer config for guest cluster nodes %#q", instanceID))

	n := customObject.GetNamespace()
	c := &corev1alpha1.DrainerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				key.InstanceIDAnnotation: instanceID,
			},
			Labels: map[string]string{
				key.ClusterIDLabel: key.ClusterID(customObject),
			},
			Name: privateDNS,
		},
		Spec: corev1alpha1.DrainerConfigSpec{
			Guest: corev1alpha1.DrainerConfigSpecGuest{
				Cluster: corev1alpha1.DrainerConfigSpecGuestCluster{
					API: corev1alpha1.DrainerConfigSpecGuestClusterAPI{
						Endpoint: key.ClusterAPIEndpoint(customObject),
					},
					ID: key.ClusterID(customObject),
				},
				Node: corev1alpha1.DrainerConfigSpecGuestNode{
					Name: privateDNS,
				},
			},
			VersionBundle: corev1alpha1.DrainerConfigSpecVersionBundle{
				Version: "0.1.0",
			},
		},
	}

	_, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Create(c)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created drainer config for guest cluster node %#q", instanceID))
	return nil
}

func (r *Resource) privateDNSForInstance(ctx context.Context, instanceID string) (string, error) {
	i := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	controllerCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	o, err := controllerCtx.AWSClient.EC2.DescribeInstances(i)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if len(o.Reservations) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected 1 reservation, got %d", len(o.Reservations))
	}
	if len(o.Reservations[0].Instances) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected 1 node, got %d", len(o.Reservations[0].Instances))
	}

	privateDNS := *o.Reservations[0].Instances[0].PrivateDnsName

	return privateDNS, nil
}
