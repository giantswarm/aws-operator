package cloudformation

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/aws-operator/service/controller/v3/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateStackInput, err := toUpdateStackInput(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	stackName := updateStackInput.StackName
	if stackName != nil && *stackName != "" {
		_, err := r.Clients.CloudFormation.UpdateStack(&updateStackInput)
		if err != nil {
			return microerror.Maskf(err, "updating AWS cloudformation stack")
		}

		r.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: updated")
	} else {
		r.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: no need to update")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	currentStackState, err := toStackState(currentState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the main stack should be updated")

	updateState := cloudformation.UpdateStackInput{}

	if currentStackState.Name != "" && !reflect.DeepEqual(desiredStackState, currentStackState) {
		r.logger.LogCtx(ctx, "debug", "main stack should be updated")

		mainTemplate, err := r.getMainGuestTemplateBody(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		updateState.StackName = aws.String(desiredStackState.Name)
		updateState.TemplateBody = aws.String(mainTemplate)
		// CAPABILITY_NAMED_IAM is required for updating IAM roles (worker policy)
		updateState.Capabilities = []*string{
			aws.String(namedIAMCapability),
		}
	}

	return updateState, nil
}
