package workerasgname

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v24/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

const (
	Name = "workerasgnamev24"
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

// EnsureCreated retrieves worker ASG name from CF stack when it is ready.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var customObject v1alpha1.AWSConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "fetching latest version of custom resource")

		oldObj, err := key.ToCustomObject(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		newObj, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(oldObj.GetNamespace()).Get(oldObj.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		customObject = *newObj

		r.logger.LogCtx(ctx, "level", "debug", "message", "fetched latest version of custom resource")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster worker ASG name in the CR")

		if customObject.Status.AWS.AutoScalingGroup.Name != "" {
			cc.Status.Drainer.WorkerASGName = customObject.Status.AWS.AutoScalingGroup.Name

			r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster worker ASG name in the CR")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster worker ASG name in the CR")
	}

	var workerASGName string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the guest cluster worker ASG name in the cloud formation stack")

		stackOutputs, stackStatus, err := cc.CloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(customObject))
		if cloudformationservice.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster worker ASG name in the cloud formation stack")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack is not yet created")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if cloudformationservice.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster worker ASG name in the cloud formation stack")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the guest cluster main stack output values are not accessible due to stack status '%s'", stackStatus))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		workerASGName, err = cc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerASGKey)
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
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack is not upgraded to the newest version yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster worker ASG name in the cloud formation stack")
	}

	// Store the ASG name in the CR status.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR status")

		customObject.Status.AWS.AutoScalingGroup.Name = workerASGName

		_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(customObject.Namespace).UpdateStatus(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated CR status")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
