package drainerinitializer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/asg"
)

const (
	Name = "drainerinitializer"
)

type ResourceConfig struct {
	ASG       asg.Interface
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	LabelMapFunc      func(cr metav1.Object) map[string]string
	LifeCycleHookName string
	ToClusterFunc     func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

type Resource struct {
	asg       asg.Interface
	g8sClient versioned.Interface
	logger    micrologger.Logger

	labelMapFunc      func(cr metav1.Object) map[string]string
	lifeCycleHookName string
	toClusterFunc     func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.ASG == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ASG must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.LabelMapFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.LabelMapFunc must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}
	if config.LifeCycleHookName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.LifeCycleHookName must not be empty", config)
	}

	r := &Resource{
		asg:       config.ASG,
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		labelMapFunc:      config.LabelMapFunc,
		toClusterFunc:     config.ToClusterFunc,
		lifeCycleHookName: config.LifeCycleHookName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) createDrainerConfig(ctx context.Context, cl infrastructurev1alpha2.AWSCluster, cr metav1.Object, instanceID, privateDNS string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating drainer config for ec2 instance %#q", instanceID))

	dc := &g8sv1alpha1.DrainerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotation.InstanceID: instanceID,
			},
			Labels: r.labelMapFunc(cr),
			Name:   privateDNS,
		},
		Spec: g8sv1alpha1.DrainerConfigSpec{
			Guest: g8sv1alpha1.DrainerConfigSpecGuest{
				Cluster: g8sv1alpha1.DrainerConfigSpecGuestCluster{
					API: g8sv1alpha1.DrainerConfigSpecGuestClusterAPI{
						Endpoint: key.ClusterAPIEndpoint(cl),
					},
					ID: key.ClusterID(cr),
				},
				Node: g8sv1alpha1.DrainerConfigSpecGuestNode{
					Name: privateDNS,
				},
			},
			VersionBundle: g8sv1alpha1.DrainerConfigSpecVersionBundle{
				Version: "0.2.0",
			},
		},
	}

	_, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(cr.GetNamespace()).Create(dc)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created drainer config for ec2 instance %#q", instanceID))

	return nil
}

// ensure creates DrainerConfigs for ASG instances in terminating/wait state
// then lets node-operator to do its job.
func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cl, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var asgName string
	{
		drainable, err := r.asg.Drainable(ctx, cr)
		if asg.IsNoASG(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find any auto scaling group")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if asg.IsNoDrainable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find any drainable auto scaling group yet")

			if key.IsDeleted(cr) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		asgName = drainable
	}

	var instances []*autoscaling.Instance
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding ec2 instances in %#q state", autoscaling.LifecycleStateTerminatingWait))

		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String(asgName),
			},
		}

		o, err := cc.Client.TenantCluster.AWS.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		var c int
		for _, g := range o.AutoScalingGroups {
			for _, i := range g.Instances {
				c++
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking instance %#q with state %#q", *i.InstanceId, *i.LifecycleState))

				if *i.LifecycleState == autoscaling.LifecycleStateTerminatingWait || *i.LifecycleState == autoscaling.LifecycleStateTerminatingProceed {
					instances = append(instances, i)
				}
			}
		}

		// In case there aren't any EC2 instances in the ASG we assume all draining
		// and deletion is properly done.
		if c == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find any ec2 instance")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if len(instances) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find ec2 instances in %#q state", autoscaling.LifecycleStateTerminatingWait))

			// In case there aren't EC2 instances in Terminating:Wait state, we cancel
			// and keep finalizers on delete events, so we try again on the next
			// reconciliation loop.
			if key.IsDeleted(cr) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d ec2 instances in %#q state", len(instances), autoscaling.LifecycleStateTerminatingWait))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring drainer configs for %d ec2 instances in %#q state", len(instances), autoscaling.LifecycleStateTerminatingWait))

		for _, instance := range instances {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding drainer config for ec2 instance %#q", *instance.InstanceId))

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
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no private DNS for ec2 instance %#q", *instance.InstanceId))
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not draining ec2 instance %#q", *instance.InstanceId))

				// Terminated instance that still have lifecycle action in the AWS API.
				// Lets finish lifecycle hook to get rid of the instance in next loop.
				err = r.completeLifeCycleHook(ctx, *instance.InstanceId, asgName)
				if err != nil {
					return microerror.Mask(err)
				}
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook for terminated ec2 instance %#q", *instance.InstanceId))
				continue
			}

			_, err = r.g8sClient.CoreV1alpha1().DrainerConfigs(cr.GetNamespace()).Get(privateDNS, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find drainer config for ec2 instance %#q", *instance.InstanceId))

				err := r.createDrainerConfig(ctx, cl, cr, *instance.InstanceId, privateDNS)
				if err != nil {
					return microerror.Mask(err)
				}

			} else if err != nil {
				return microerror.Mask(err)
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found drainer config for ec2 instance %#q", *instance.InstanceId))
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured drainer configs for %d ec2 instances in %#q state", len(instances), autoscaling.LifecycleStateTerminatingWait))
	}

	return nil
}

func (r *Resource) privateDNSForInstance(ctx context.Context, instanceID string) (string, error) {
	i := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
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

func (r *Resource) completeLifeCycleHook(ctx context.Context, instanceID, asgName string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completing life cycle hook action for tenant cluster node %#q", instanceID))
	i := &autoscaling.CompleteLifecycleActionInput{
		AutoScalingGroupName:  aws.String(asgName),
		InstanceId:            aws.String(instanceID),
		LifecycleActionResult: aws.String("CONTINUE"),
		LifecycleHookName:     aws.String(r.lifeCycleHookName),
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = cc.Client.TenantCluster.AWS.AutoScaling.CompleteLifecycleAction(i)
	if IsNoActiveLifeCycleAction(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not found life cycle hook action for tenant cluster node %#q", instanceID))
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed life cycle hook action for tenant cluster node %#q", instanceID))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed life cycle hook action for tenant cluster node %#q", instanceID))

	return nil
}
