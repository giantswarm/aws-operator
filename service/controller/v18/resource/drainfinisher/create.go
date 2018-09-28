package drainfinisher

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/v17/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v17/key"
)

// EnsureCreated completes ASG lifecycle hooks for nodes drained by
// node-operator, and then deletes drained DrainerConfigs.
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

	var drainedDrainerConfigs []corev1alpha1.DrainerConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding drained drainer configs for the guest cluster")

		n := v1.NamespaceAll
		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", key.ClusterIDLabel, key.ClusterID(customObject)),
		}

		drainerConfigs, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, drainerConfig := range drainerConfigs.Items {
			if drainerConfig.Status.HasDrainedCondition() {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("drainer config '%s' of guest cluster has drained condition", drainerConfig.GetName()))
				drainedDrainerConfigs = append(drainedDrainerConfigs, drainerConfig)
			}

			if drainerConfig.Status.HasTimeoutCondition() {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("drainer config '%s' of guest cluster has timeout condition", drainerConfig.GetName()))
				drainedDrainerConfigs = append(drainedDrainerConfigs, drainerConfig)
			}
		}

		if len(drainedDrainerConfigs) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find drained drainer configs for the guest cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d drained drainer configs for the guest cluster", len(drainedDrainerConfigs)))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring finised draining for drained nodes")

		for _, drainerConfig := range drainedDrainerConfigs {
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
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured finised draining for drained nodes")
	}

	return nil
}

func (r *Resource) completeLifecycleHook(ctx context.Context, instanceID, workerASGName string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completing lifecycle hook action for guest cluster node '%s'", instanceID))
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
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not found lifecycle hook action for guest cluster node '%s'", instanceID))
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook action for guest cluster node '%s'", instanceID))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook action for guest cluster node '%s'", instanceID))
	return nil
}

func (r *Resource) deleteDrainerConfig(ctx context.Context, drainerConfig corev1alpha1.DrainerConfig) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting drainer config for guest cluster node '%s'", drainerConfig.Name))

	n := drainerConfig.GetNamespace()
	i := drainerConfig.GetName()
	o := &metav1.DeleteOptions{}

	err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Delete(i, o)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted drainer config for guest cluster node '%s'", drainerConfig.Name))
	return nil
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
