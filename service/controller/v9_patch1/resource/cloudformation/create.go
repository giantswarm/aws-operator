package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v9_patch1/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cluster, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	stackInput, err := toCreateStackInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackInput.StackName != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the guest cluster main stack")

		_, err = r.clients.CloudFormation.CreateStack(&stackInput)
		if err != nil {
			return microerror.Mask(err)
		}
		err = r.clients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: stackInput.StackName,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		// Create host post-main stack. It includes the peering routes, which need resources from the
		// guest stack to be in place before it can be created.
		err = r.createHostPostStack(ctx, cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the guest cluster main stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not creating the guest cluster main stack")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}
	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the guest cluster main stack has to be created")

	createState := cloudformation.CreateStackInput{}

	if currentStackState.Name == "" || desiredStackState.Name != currentStackState.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be created")

		if err := r.validateCluster(customObject); err != nil {
			return cloudformation.CreateStackInput{}, microerror.Mask(err)
		}

		// We need to create the required peering resources in the host account before
		// getting the guest main stack template body, it requires id values from host
		// resources.
		err = r.createHostPreStack(ctx, customObject)
		if err != nil {
			return cloudformation.CreateStackInput{}, microerror.Mask(err)
		}

		var mainTemplate string
		mainTemplate, err := r.getMainGuestTemplateBody(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		createState.StackName = aws.String(desiredStackState.Name)
		createState.TemplateBody = aws.String(mainTemplate)
		createState.TimeoutInMinutes = aws.Int64(defaultCreationTimeout)
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		createState.Capabilities = []*string{
			aws.String(namedIAMCapability),
		}

		createState.SetTags(r.getCloudFormationTags(customObject))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not have to be created")
	}

	return createState, nil
}

func (r *Resource) createHostPreStack(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	stackName := key.MainHostPreStackName(customObject)
	mainTemplate, err := r.getMainHostPreTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		Capabilities: []*string{
			aws.String(namedIAMCapability),
		},
	}
	createStack.SetTags(r.getCloudFormationTags(customObject))

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating the host cluster pre cloud formation stack")

	_, err = r.hostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.hostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created the host cluster pre cloud formation stack")

	return nil
}

func (r *Resource) createHostPostStack(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	stackName := key.MainHostPostStackName(customObject)
	mainTemplate, err := r.getMainHostPostTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
	}
	createStack.SetTags(r.getCloudFormationTags(customObject))

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating the host cluster post cloud formation stack")

	_, err = r.hostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.hostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created the host cluster post cloud formation stack")

	return nil
}
