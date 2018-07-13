package lifecycle

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v14/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v14/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v14/key"
)

// EnsureCreated tries to drain guest cluster nodes when necessary. Once it
// detects a guest cluster node being in terminating/wait state in EC2 a
// DrainerConfig is created to instruct the node-operator to drain the specific
// node. The node-operator updates the DrainerConfig state as soon as it has
// drained the node and the aws-operator here completes the lifecycle hook of
// the drained node and deletes the DrainerConfig.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var stackOutputs []*cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the guest cluster main stack in the AWS API")

		stackOutputs, _, err = sc.CloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(customObject))
		if cloudformationservice.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster main stack in the AWS API")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil

		} else if cloudformationservice.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack output values are not accessible due to stack state transition")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster main stack in the AWS API")
	}

	var workerASGName string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the guest cluster worker ASG name in the cloud formation stack")

		workerASGName, err = sc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerASGKey)
		if cloudformationservice.IsOutputNotFound(err) {
			// Since we are transitioning between versions we will have situations in
			// which old clusters are updated to new versions and miss the ASG name in
			// the CF stack outputs. We stop resource reconciliation at this point to
			// reconcile again at a later point when the stack got upgraded and
			// contains the ASG name.
			//
			// TODO remove this condition as soon as all guest clusters in existence
			// obtain a ASG name.
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster worker ASG name in the cloud formation stack")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster worker ASG name in the cloud formation stack")
	}

	var instances []*autoscaling.Instance
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("looking for the guest cluster instances being in state '%s'", autoscaling.LifecycleStateTerminatingWait))

		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String(workerASGName),
			},
		}

		o, err := sc.AWSClient.AutoScaling.DescribeAutoScalingGroups(i)
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
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the guest cluster instances being in state '%s'", autoscaling.LifecycleStateTerminatingWait))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d guest cluster instances being in state '%s'", len(instances), autoscaling.LifecycleStateTerminatingWait))
		}
	}

	{
		for _, instance := range instances {
			r.logger.LogCtx(ctx, "level", "debug", "message", "looking for drainer config for the guest cluster")

			privateDNS, err := r.privateDNSForInstance(ctx, *instance.InstanceId)
			if err != nil {
				return microerror.Mask(err)
			}

			n := customObject.GetNamespace()
			o := metav1.GetOptions{}

			_, err = r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Get(privateDNS, o)
			if errors.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "did not find drainer config for guest cluster node")

				err := r.createDrainerConfig(ctx, customObject, *instance.InstanceId, privateDNS)
				if err != nil {
					return microerror.Mask(err)
				}

			} else if err != nil {
				return microerror.Mask(err)
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "found drainer config for the guest cluster")

				r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for inspection of drainer config for the guest cluster")
			}
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "start inspection of drainer configs for the guest cluster")

		n := v1.NamespaceAll
		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", key.ClusterIDLabel, key.ClusterID(customObject)),
		}

		drainerConfigs, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(drainerConfigs.Items) > 0 {
			for _, drainerConfig := range drainerConfigs.Items {
				r.logger.LogCtx(ctx, "level", "debug", "message", "inspecting drainer config for the guest cluster")

				if drainerConfig.Status.HasDrainedCondition() {
					r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config of guest cluster has drained condition")

					err := r.finishDraining(ctx, drainerConfig, workerASGName)
					if err != nil {
						return microerror.Mask(err)
					}
				}

				if drainerConfig.Status.HasTimeoutCondition() {
					r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config of guest cluster has timeout condition")

					err := r.finishDraining(ctx, drainerConfig, workerASGName)
					if err != nil {
						return microerror.Mask(err)
					}
				}

				r.logger.LogCtx(ctx, "level", "debug", "message", "inspected drainer config for the guest cluster")
			}
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "no drainer configs to inspect for the guest cluster")
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "finished inspection of drainer configs for the guest cluster")
	}

	return nil
}

func (r *Resource) completeLifecycleHook(ctx context.Context, instanceID, workerASGName string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completing lifecycle hook action for guest cluster instance '%s'", instanceID))

	i := &autoscaling.CompleteLifecycleActionInput{
		AutoScalingGroupName:  aws.String(workerASGName),
		InstanceId:            aws.String(instanceID),
		LifecycleActionResult: aws.String("CONTINUE"),
		LifecycleHookName:     aws.String(key.NodeDrainerLifecycleHookName),
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = sc.AWSClient.AutoScaling.CompleteLifecycleAction(i)
	if IsNoActiveLifecycleAction(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no lifecycle hook action for guest cluster instance '%s'", instanceID))
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook action for guest cluster instance '%s'", instanceID))
	}

	return nil
}

func (r *Resource) createDrainerConfig(ctx context.Context, customObject providerv1alpha1.AWSConfig, instanceID, privateDNS string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "creating drainer config for guest cluster node")

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

	r.logger.LogCtx(ctx, "level", "debug", "message", "created drainer config for guest cluster node")

	return nil
}

func (r *Resource) deleteDrainerConfig(ctx context.Context, drainerConfig corev1alpha1.DrainerConfig) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting drainer config for guest cluster node")

	n := drainerConfig.GetNamespace()
	i := drainerConfig.GetName()
	o := &metav1.DeleteOptions{}

	err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Delete(i, o)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleted drainer config for guest cluster node")

	return nil
}

func (r *Resource) finishDraining(ctx context.Context, drainerConfig corev1alpha1.DrainerConfig, workerASGName string) error {
	// This is a special thing for AWS. We use annotations to transport EC2
	// instance IDs. Otherwise the lookups of all necessary information
	// again would be quite a ball ache. Se we take the shortcut leveraging
	// the k8s API.
	instanceID, err := instanceIDFromAnnotations(drainerConfig.GetAnnotations())
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.completeLifecycleHook(ctx, instanceID, workerASGName)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.deleteDrainerConfig(ctx, drainerConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) privateDNSForInstance(ctx context.Context, instanceID string) (string, error) {
	i := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	o, err := sc.AWSClient.EC2.DescribeInstances(i)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if len(o.Reservations) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected 1 reservation, got %d", len(o.Reservations))
	}
	if len(o.Reservations[0].Instances) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected 1 instance, got %d", len(o.Reservations[0].Instances))
	}

	privateDNS := *o.Reservations[0].Instances[0].PrivateDnsName

	return privateDNS, nil
}

func instanceIDFromAnnotations(annotations map[string]string) (string, error) {
	instanceID, ok := annotations[key.InstanceIDAnnotation]
	if !ok {
		return "", microerror.Maskf(missingAnnotationError, key.InstanceIDAnnotation)
	}
	if instanceID == "" {
		return "", microerror.Maskf(missingAnnotationError, key.InstanceIDAnnotation)
	}

	return instanceID, nil
}
