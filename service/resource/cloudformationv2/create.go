package cloudformationv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cluster, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	stackInput, err := toCreateStackInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.Clients.CloudFormation.CreateStack(&stackInput)
	if err != nil {
		return microerror.Mask(err)
	}
	err = r.HostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: stackInput.StackName,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "creating AWS cloudformation stack: created")

	// Create host post-main stack. It includes the peering routes, which need resources from the
	// guest stack to be in place before it can be created.
	err = r.createHostPostStack(cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if the main stack should be created")

	createState := cloudformation.CreateStackInput{
		StackName: aws.String(""),
	}

	if desiredStackState.Name != "" {
		r.logger.LogCtx(ctx, "debug", "main stack should be created")

		// We need to create the required peering resources in the host account before
		// getting the guest main stack template body, it requires id values from host
		// resources.
		err = r.createHostPreStack(customObject)
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
			aws.String("CAPABILITY_NAMED_IAM"),
		}
	}

	return createState, nil
}
