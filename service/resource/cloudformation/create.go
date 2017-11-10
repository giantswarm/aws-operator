package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	stackInput, ok := createChange.(awscloudformation.CreateStackInput)
	if !ok {
		return microerror.Mask(fmt.Errorf("unexpected type, expecting %T, got %T", stackInput, createChange))
	}

	_, err := r.awsClient.CreateStack(&stackInput)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return awscloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	currentStackState, ok := currentState.(StackState)
	if !ok {
		return awscloudformation.CreateStackInput{}, microerror.Mask(fmt.Errorf("unexpected type, expecting %T, got %T", currentStackState, currentState))
	}
	desiredStackState, ok := desiredState.(StackState)
	if !ok {
		return awscloudformation.CreateStackInput{}, microerror.Mask(fmt.Errorf("unexpected type, expecting %T, got %T", desiredStackState, desiredState))
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out if the main stack should be created")

	createState := awscloudformation.CreateStackInput{
		StackName: aws.String(""),
	}

	if currentStackState.Name != desiredStackState.Name {
		var mainTemplate string
		/*
			      commented out until we assing proper values to the template
						mainTemplate, err := getMainTemplateBody(customObject)
						if err != nil {
							return nil, microerror.Mask(err)
						}
		*/
		createState.StackName = aws.String(desiredStackState.Name)
		createState.TemplateBody = aws.String(mainTemplate)
		createState.TimeoutInMinutes = aws.Int64(defaultCreationTimeout)
	}

	return createState, nil
}
