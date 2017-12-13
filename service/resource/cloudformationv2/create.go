package cloudformationv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	stackInput, err := toCreateStackInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.awsClients.CloudFormation.CreateStack(&stackInput)
	if err != nil {
		return err
	}

	r.logger.LogCtx(ctx, "debug", "creating AWS cloudformation stack: created")

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return awscloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return awscloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the main stack should be created")

	createState := awscloudformation.CreateStackInput{
		StackName: aws.String(""),
	}

	if desiredStackState.Name != "" {
		var mainTemplate string
		mainTemplate, err := r.getMainTemplateBody(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		createState.StackName = aws.String(desiredStackState.Name)
		createState.TemplateBody = aws.String(mainTemplate)
		createState.TimeoutInMinutes = aws.Int64(defaultCreationTimeout)
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		createState.Capabilities = []*string{
			aws.String("CAPABILITY_NAMED_IAM"),
		}
	}

	return createState, nil
}
