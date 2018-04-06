package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	stackStateToDelete, err := toStackState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackStateToDelete.Status == cloudformation.ResourceStatusUpdateInProgress {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cannot delete CF stacks due to stack state transitioning")

		// TODO control flow via more proper mechanism via something like the
		// context like it is done for cancelation already.
		return microerror.Maskf(deletionMustBeRetriedError, "stack state is transitioning")
	}

	if stackStateToDelete.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the guest cluster main stack")

		i := &cloudformation.DeleteStackInput{
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

	if stackStateToDelete.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the host cluster pre stack")

		i := &cloudformation.DeleteStackInput{
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

	if stackStateToDelete.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the host cluster post stack")

		i := &cloudformation.DeleteStackInput{
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
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(deleteChange)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	// For deletion we need the current and desired state in order to process it
	// reliable. There are cases in which the current state does not contain a
	// name because of stack state transitions but provides the stack status
	// itself. The desired state never provides the current stack state in return.
	// This is why we have to compute the delete state using both entities. Note
	// that for deletion some properties of the stack state are omitted due to
	// unimportance to the process.
	deleteState := StackState{
		Name:   desiredStackState.Name,
		Status: currentStackState.Status,
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster main stack that has to be deleted")

	return deleteState, nil
}
