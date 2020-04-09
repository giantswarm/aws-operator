package drainfinisher

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "drainfinisher"
)

type ResourceConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) completeLifecycleHook(ctx context.Context, instanceID, workerASGName string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completing lifecycle hook action for tenant cluster node %#q", instanceID))
	i := &autoscaling.CompleteLifecycleActionInput{
		AutoScalingGroupName:  aws.String(workerASGName),
		InstanceId:            aws.String(instanceID),
		LifecycleActionResult: aws.String("CONTINUE"),
		LifecycleHookName:     aws.String("NodePool"),
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = cc.Client.TenantCluster.AWS.AutoScaling.CompleteLifecycleAction(i)
	if IsNoActiveLifecycleAction(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not found lifecycle hook action for tenant cluster node %#q", instanceID))
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook action for tenant cluster node %#q", instanceID))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook action for tenant cluster node %#q", instanceID))
	return nil
}

func (r *Resource) deleteDrainerConfig(ctx context.Context, drainerConfig corev1alpha1.DrainerConfig) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting drainer config for tenant cluster node %#q", drainerConfig.Name))

	n := drainerConfig.GetNamespace()
	i := drainerConfig.GetName()
	o := &metav1.DeleteOptions{}

	err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Delete(i, o)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted drainer config for tenant cluster node %#q", drainerConfig.Name))
	return nil
}

// ensure completes ASG lifecycle hooks for nodes drained by node-operator, and
// then deletes drained DrainerConfigs.
func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	workerASGName := cc.Status.TenantCluster.ASG.Name
	if workerASGName == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "worker ASG name is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var drainedDrainerConfigs []corev1alpha1.DrainerConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding drained drainer configs for tenant cluster")

		n := cr.GetNamespace()
		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s, %s=%s", label.Cluster, key.ClusterID(&cr), label.MachineDeployment, key.MachineDeploymentID(&cr)),
		}

		drainerConfigs, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		// As long as we have DrainerConfig CRs for this tenant cluster we want to
		// keep finalizers and try again next time. We keep finalizers because this
		// might be a delete event.
		if len(drainerConfigs.Items) != 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found drainer configs for tenant cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)
		}

		for _, drainerConfig := range drainerConfigs.Items {
			if drainerConfig.Status.HasDrainedCondition() {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("drainer config %#q of tenant cluster has drained condition", drainerConfig.GetName()))
				drainedDrainerConfigs = append(drainedDrainerConfigs, drainerConfig)
			}

			if drainerConfig.Status.HasTimeoutCondition() {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("drainer config %#q of tenant cluster has timeout condition", drainerConfig.GetName()))
				drainedDrainerConfigs = append(drainedDrainerConfigs, drainerConfig)
			}
		}

		if len(drainedDrainerConfigs) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find drained drainer configs for tenant cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d drained drainer configs for tenant cluster", len(drainedDrainerConfigs)))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring finished draining for drained nodes")

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

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured finished draining for drained nodes")
	}

	return nil
}

func instanceIDFromAnnotations(annotations map[string]string) (string, error) {
	instanceID, ok := annotations[annotation.InstanceID]
	if !ok {
		return "", microerror.Maskf(missingAnnotationError, annotation.InstanceID)
	}
	if instanceID == "" {
		return "", microerror.Maskf(missingAnnotationError, annotation.InstanceID)
	}

	return instanceID, nil
}
