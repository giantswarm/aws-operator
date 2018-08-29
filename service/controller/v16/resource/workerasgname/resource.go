package workerasgname

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v15/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v15/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

const (
	Name = "workerasgnamev15"
)

type ResourceConfig struct {
	Logger micrologger.Logger
}

type Resource struct {
	logger micrologger.Logger
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// EnsureCreated retrieves worker ASG name from CF stack when it is ready.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
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

		stackOutputs, stackStatus, err := controllerCtx.CloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(customObject))
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

	controllerCtx.Status.Drainer.WorkerASGName = workerASGName

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
