package drainerfinalizer

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	corev1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/finalizerskeptcontext"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/aws-operator/v14/pkg/annotation"
	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
	"github.com/giantswarm/aws-operator/v14/service/internal/asg"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "drainerfinalizer"
)

type ResourceConfig struct {
	ASG        asg.Interface
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger

	LabelMapFunc      func(cr metav1.Object) map[string]string
	LifeCycleHookName string
}

type Resource struct {
	asg        asg.Interface
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger

	labelMapFunc      func(cr metav1.Object) map[string]string
	lifeCycleHookName string
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.ASG == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ASG must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.LabelMapFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.LabelMapFunc must not be empty", config)
	}
	if config.LifeCycleHookName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.LifeCycleHookName must not be empty", config)
	}

	r := &Resource{
		asg:        config.ASG,
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,

		labelMapFunc:      config.LabelMapFunc,
		lifeCycleHookName: config.LifeCycleHookName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) completeLifeCycleHook(ctx context.Context, instanceID, asgName string) error {
	r.logger.Debugf(ctx, "completing life cycle hook action for tenant cluster node %#q", instanceID)

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
		r.logger.Debugf(ctx, "did not find life cycle hook action for tenant cluster node %#q", instanceID)
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "completed life cycle hook action for tenant cluster node %#q", instanceID)

	return nil
}

func (r *Resource) deleteDrainerConfig(ctx context.Context, dc corev1alpha1.DrainerConfig) error {
	r.logger.Debugf(ctx, "deleting drainer config for tenant cluster node %#q", dc.Name)

	o := &metav1.DeleteOptions{}

	err := r.ctrlClient.Delete(ctx, &dc, &ctrlClient.DeleteOptions{Raw: o})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleted drainer config for tenant cluster node %#q", dc.Name)
	return nil
}

// ensure completes ASG life cycle hooks for nodes drained by node-operator, and
// then deletes drained DrainerConfigs.
func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var asgName string
	{
		drainable, err := r.asg.Drainable(ctx, cr)
		if asg.IsNoASG(err) {
			r.logger.Debugf(ctx, "did not find any auto scaling group")
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if asg.IsNoDrainable(err) {
			r.logger.Debugf(ctx, "did not find any drainable auto scaling group yet")

			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		asgName = drainable
	}

	drainerConfigs := &corev1alpha1.DrainerConfigList{}
	{
		r.logger.Debugf(ctx, "finding drained drainer configs for tenant cluster")

		n := cr.GetNamespace()
		o := metav1.ListOptions{
			LabelSelector: labels.Set(r.labelMapFunc(cr)).String(),
		}

		err := r.ctrlClient.List(ctx, drainerConfigs, &ctrlClient.ListOptions{Raw: &o, Namespace: n})
		if err != nil {
			return microerror.Mask(err)
		}

		// As long as we have DrainerConfig CRs for this tenant cluster we want to
		// keep finalizers and try again next time. We keep finalizers because this
		// might be a delete event.
		if len(drainerConfigs.Items) != 0 {
			r.logger.Debugf(ctx, "found drainer configs for tenant cluster")

			if key.IsDeleted(cr) {
				r.logger.Debugf(ctx, "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)
			}
		}

		for _, dc := range drainerConfigs.Items {
			if dc.Status.HasDrainedCondition() {
				r.logger.Debugf(ctx, "drainer config %#q of tenant cluster has drained condition", dc.GetName())
			}

			if dc.Status.HasTimeoutCondition() {
				r.logger.Debugf(ctx, "drainer config %#q of tenant cluster has timeout condition", dc.GetName())
			}
		}

		if len(drainerConfigs.Items) == 0 {
			r.logger.Debugf(ctx, "did not find any drainer config for tenant cluster")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		r.logger.Debugf(ctx, "found %d drained drainer configs for tenant cluster", len(drainerConfigs.Items))
	}

	{
		r.logger.Debugf(ctx, "ensuring finished draining for drained nodes")

		for _, dc := range drainerConfigs.Items {
			if dc.Status.HasDrainedCondition() || dc.Status.HasTimeoutCondition() || key.IsDeleted(cr) {
				// This is a special thing for AWS. We use annotations to transport EC2
				// instance IDs. Otherwise the lookups of all necessary information
				// again would be quite a ball ache. Se we take the shortcut leveraging
				// the k8s API.
				instanceID, err := instanceIDFromAnnotations(dc.GetAnnotations())
				if err != nil {
					return microerror.Mask(err)
				}

				if asgName != "" {
					err = r.completeLifeCycleHook(ctx, instanceID, asgName)
					if err != nil {
						return microerror.Mask(err)
					}
				}

				err = r.deleteDrainerConfig(ctx, dc)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}

		r.logger.Debugf(ctx, "ensured finished draining for drained nodes")
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
