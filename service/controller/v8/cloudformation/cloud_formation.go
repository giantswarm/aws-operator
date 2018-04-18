package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

type Config struct {
	Client *cloudformation.CloudFormation
}

type CloudFormation struct {
	client *cloudformation.CloudFormation
}

func New(config Config) (*CloudFormation, error) {
	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}

	c := &CloudFormation{
		client: config.Client,
	}

	return c, nil
}

func (c *CloudFormation) DescribeOutputsAndStatus(stackName string) ([]*cloudformation.Output, string, error) {
	// At first we fetch the CF stack state by describing it via the AWS golang
	// SDK. We are interested in the stack outputs and the stack status, since
	// this tells us if we are able to access outputs at all.
	var stackOutputs []*cloudformation.Output
	var stackStatus string
	{
		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}

		o, err := c.client.DescribeStacks(i)
		if IsStackNotFound(err) {
			return nil, "", microerror.Maskf(stackNotFoundError, "stack name '%s'", stackName)
		} else if err != nil {
			return nil, "", microerror.Mask(err)
		}

		if len(o.Stacks) > 1 {
			return nil, "", microerror.Maskf(tooManyStacksError, "expected 1 stack, got %d", len(o.Stacks))
		}

		stackOutputs = o.Stacks[0].Outputs
		stackStatus = *o.Stacks[0].StackStatus
	}

	// We call DescribeOutputsAndStatus in certain GetCurrentState
	// implementations. If the stack creation failed, the outputs can be
	// unaccessible. This can lead to a stack that cannot be deleted. it can also
	// be called during creation, while the outputs are still not accessible.
	{
		errorStatuses := []string{
			cloudformation.StackStatusRollbackInProgress,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusCreateInProgress,
		}

		for _, s := range errorStatuses {
			if stackStatus == s {
				return nil, "", microerror.Maskf(outputsNotAccessibleError, "due to stack state '%s'", stackStatus)
			}
		}
	}

	return stackOutputs, stackStatus, nil
}

func (c *CloudFormation) GetOutputValue(outputs []*cloudformation.Output, key string) (string, error) {
	for _, o := range outputs {
		if *o.OutputKey == key {
			return *o.OutputValue, nil
		}
	}

	return "", microerror.Maskf(outputNotFoundError, "stack output value for key '%s'", key)
}
