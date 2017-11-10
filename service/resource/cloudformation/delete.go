package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deleteStackInput, err := toDeleteStackInput(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	stackName := deleteStackInput.StackName
	if *stackName != "" {
		_, err := r.awsClient.DeleteStack(&deleteStackInput)

		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "deleted stacks from the AWS API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the stacks do not need to be deleted from the AWS API")
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
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return awscloudformation.DeleteStackInput{}, microerror.Mask(err)
	}

	currentStackState, err := toStackState(currentState)
	if err != nil {
		return awscloudformation.DeleteStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return awscloudformation.DeleteStackInput{}, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out if the main stack should be deleted")

	createState := awscloudformation.DeleteStackInput{
		StackName: aws.String(""),
	}

	if desiredStackState.Name != currentStackState.Name {
		createState.StackName = aws.String(currentStackState.Name)
	}

	return createState, nil
}
