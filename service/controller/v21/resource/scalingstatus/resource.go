package scalingstatus

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v21/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v21/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v21/key"
)

const (
	Name = "scalingstatusv21"
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

// EnsureCreated retrieves worker ASG Desired value when it is ready and writes
// it to status field.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	controllerCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var workerASGName string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the guest cluster worker ASG name in the cloud formation stack")

		stackOutputs, stackStatus, err := controllerCtx.CloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(customResource))
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

		workerASGName, err = controllerCtx.CloudFormation.GetOutputValue(stackOutputs, key.WorkerASGKey)
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

	var desiredCapacity int
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out desired value in worker asg")

		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				&workerASGName,
			},
		}
		o, err := controllerCtx.AWSClient.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.AutoScalingGroups) == 0 {
			return microerror.Maskf(notFoundError, "asg for name %s", workerASGName)
		}

		if o.AutoScalingGroups[0].DesiredCapacity == nil {
			return microerror.Maskf(notFoundError, "desired capacity for asg is nil")
		}

		desiredCapacity = int(*o.AutoScalingGroups[0].DesiredCapacity)

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found out desired value in worker asg: %d", desiredCapacity))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "fetching latest version of custom resource")

		newObj, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(customResource.GetNamespace()).Get(customResource.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		customResource = *newObj

		r.logger.LogCtx(ctx, "level", "debug", "message", "fetched latest version of custom resource")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR status")

		customResource.Status.Cluster.Scaling.DesiredCapacity = desiredCapacity

		_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(customResource.Namespace).UpdateStatus(&customResource)
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
