package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/aws-operator/service/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	// no-op if we are not using cloudformation
	if !key.UseCloudFormation(customObject) {
		r.logger.LogCtx(ctx, "debug", "not processing cloudformation stack")
		return nil
	}

	deleteStackInput, err := toDeleteStackInput(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	stackName := deleteStackInput.StackName
	if *stackName != "" {
		_, err := r.awsClients.CloudFormation.DeleteStack(&deleteStackInput)
		if err != nil {
			return microerror.Maskf(err, "deleting AWS CloudFormation Stack")
		}

		r.logger.LogCtx(ctx, "debug", "deleting AWS CloudFormation stack: deleted")
	} else {
		r.logger.LogCtx(ctx, "debug", "deleting AWS CloudFormation stack: already deleted")
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
		return awscloudformation.DeleteStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return awscloudformation.DeleteStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the main stack should be deleted")

	deleteState := awscloudformation.DeleteStackInput{
		StackName: aws.String(""),
	}

	if currentStackState.Name != "" && desiredStackState.Name != currentStackState.Name {
		deleteState.StackName = aws.String(currentStackState.Name)
	}

	return deleteState, nil
}
