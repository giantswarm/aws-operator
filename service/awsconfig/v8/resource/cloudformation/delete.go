package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v8/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	stackInputToDelete, err := toStackInput(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackInputToDelete != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the guest cluster main stack")

		i := &awscloudformation.DeleteStackInput{
			StackName: aws.String(key.MainGuestStackName(customObject)),
		}
		_, err = r.clients.CloudFormation.DeleteStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the guest cluster main stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting the guest cluster main stack")
	}

	if stackInputToDelete != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the host cluster pre stack")

		i := &awscloudformation.DeleteStackInput{
			StackName: aws.String(key.MainHostPreStackName(customObject)),
		}
		_, err = r.hostClients.CloudFormation.DeleteStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the host cluster pre stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting the host cluster pre stack")
	}

	if stackInputToDelete != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the host cluster post stack")

		i := &awscloudformation.DeleteStackInput{
			StackName: aws.String(key.MainHostPostStackName(customObject)),
		}
		_, err = r.hostClients.CloudFormation.DeleteStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the host cluster post stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting the host cluster post stack")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the guest cluster main stack has to be deleted")

	var deleteState *StackState

	if currentStackState.Name != "" && desiredStackState.Name != currentStackState.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be deleted")

		deleteState = &awscloudformation.DeleteStackInput{
			StackName: aws.String(currentStackState.Name),
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not have to be deleted")
	}

	return deleteState, nil
}
