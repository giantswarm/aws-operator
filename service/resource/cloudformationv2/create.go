package cloudformationv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
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

	r.logger.LogCtx(ctx, "debug", "creating AWS cloudformation stack")

	if stackInput.StackName != nil {
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

		// Create host post-main stack. It includes the peering routes, which need resources from the
		// guest stack to be in place before it can be created.
		err = r.createHostPostStack(cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "creating AWS cloudformation stack: created")
	} else {
		r.logger.LogCtx(ctx, "debug", "creating AWS cloudformation stack: already created")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
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

	r.logger.LogCtx(ctx, "debug", "finding out if the main stack should be created")

	createState := cloudformation.CreateStackInput{}

	r.logger.Log("info", fmt.Sprintf("CURRENT STATE %#v", currentStackState))
	r.logger.Log("info", fmt.Sprintf("DESIRED STATE %#v", desiredStackState))

	if currentStackState.Name == "" || desiredStackState.Name != currentStackState.Name {
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
			aws.String(namedIAMCapability),
		}
	}

	r.logger.Log("info", fmt.Sprintf("CREATE STATE %#v", createState))

	return createState, nil
}

func (r *Resource) createHostPreStack(customObject v1alpha1.AWSConfig) error {
	stackName := keyv2.MainHostPreStackName(customObject)
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

	r.logger.Log("debug", "creating AWS Host Pre-Guest cloudformation stack")
	_, err = r.HostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.HostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Log("debug", "creating AWS Host Pre-Guest cloudformation stack: created")
	return nil
}

func (r *Resource) createHostPostStack(customObject v1alpha1.AWSConfig) error {
	stackName := keyv2.MainHostPostStackName(customObject)
	mainTemplate, err := r.getMainHostPostTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
	}

	r.logger.Log("debug", "creating AWS Host Post-Guest cloudformation stack")
	_, err = r.HostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.HostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "creating AWS Host Post-Guest cloudformation stack: created")

	return nil
}
