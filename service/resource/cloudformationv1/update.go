package cloudformationv1

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateStackInput, err := toUpdateStackInput(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	stackName := updateStackInput.StackName
	if *stackName != "" {
		_, err := r.awsClients.CloudFormation.UpdateStack(&updateStackInput)
		if err != nil {
			return microerror.Maskf(err, "updating AWS cloudformation stack")
		}

		r.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: updated")
	} else {
		r.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: no need to update")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return awscloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return awscloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	currentStackState, err := toStackState(currentState)
	if err != nil {
		return awscloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the main stack should be updated")

	updateState := awscloudformation.UpdateStackInput{
		StackName: aws.String(""),
	}

	if !reflect.DeepEqual(desiredStackState, currentStackState) {
		var mainTemplate string
		mainTemplate, err := r.getMainTemplateBody(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		updateState.StackName = aws.String(desiredStackState.Name)
		updateState.TemplateBody = aws.String(mainTemplate)
	}

	return updateState, nil
}
