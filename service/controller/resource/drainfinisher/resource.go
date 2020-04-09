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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

const (
	Name = "drainfinisher"
)

type ResourceConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	LabelSelectorFunc func(cr metav1.Object) *metav1.LabelSelector
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	labelSelectorFunc func(cr metav1.Object) *metav1.LabelSelector
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.LabelSelectorFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.LabelSelectorFunc must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		labelSelectorFunc: config.LabelSelectorFunc,
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

func (r *Resource) deleteDrainerConfig(ctx context.Context, dc corev1alpha1.DrainerConfig) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting drainer config for tenant cluster node %#q", dc.Name))

	n := dc.GetNamespace()
	i := dc.GetName()
	o := &metav1.DeleteOptions{}

	err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Delete(i, o)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted drainer config for tenant cluster node %#q", dc.Name))
	return nil
}

// ensure completes ASG lifecycle hooks for nodes drained by node-operator, and
// then deletes drained DrainerConfigs.
func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var workerASGName string
	{
		if cc.Status.TenantCluster.ASG.Name == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "worker auto scaling group name is not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		workerASGName = cc.Status.TenantCluster.ASG.Name
	}

	var drainedDrainerConfigs []corev1alpha1.DrainerConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding drained drainer configs for tenant cluster")

		n := cr.GetNamespace()
		o := metav1.ListOptions{
			LabelSelector: r.labelSelectorFunc(cr).String(),
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

		for _, dc := range drainerConfigs.Items {
			if dc.Status.HasDrainedCondition() {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("drainer config %#q of tenant cluster has drained condition", dc.GetName()))
				drainedDrainerConfigs = append(drainedDrainerConfigs, dc)
			}

			if dc.Status.HasTimeoutCondition() {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("drainer config %#q of tenant cluster has timeout condition", dc.GetName()))
				drainedDrainerConfigs = append(drainedDrainerConfigs, dc)
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

		for _, dc := range drainedDrainerConfigs {
			// This is a special thing for AWS. We use annotations to transport EC2
			// instance IDs. Otherwise the lookups of all necessary information
			// again would be quite a ball ache. Se we take the shortcut leveraging
			// the k8s API.
			instanceID, err := instanceIDFromAnnotations(dc.GetAnnotations())
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.completeLifecycleHook(ctx, instanceID, workerASGName)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.deleteDrainerConfig(ctx, dc)
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
