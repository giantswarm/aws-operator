package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsCF "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	createChangeState, ok := createChange.(StackState)
	if !ok {
		return microerror.Mask(fmt.Errorf("unexpected type, expecting %T, got %T", createChangeState, createChange))
	}

	mainTemplate, err := getMainTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	stackInput := &awsCF.CreateStackInput{
		StackName:        aws.String(createChangeState.Name),
		TemplateBody:     aws.String(mainTemplate),
		TimeoutInMinutes: aws.Int64(30),
	}

	_, err = r.awsClient.CreateStack(stackInput)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentStackState, ok := currentState.(StackState)
	if !ok {
		return nil, microerror.Mask(fmt.Errorf("unexpected type, expecting %T, got %T", currentStackState, currentState))
	}
	desiredStackState, ok := desiredState.(StackState)
	if !ok {
		return nil, microerror.Mask(fmt.Errorf("unexpected type, expecting %T, got %T", desiredStackState, desiredState))
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out if the main stack should be created")

	var createState StackState

	if currentStackState.Name != desiredStackState.Name {
		createState = desiredStackState
	}

	return createState, nil
}
