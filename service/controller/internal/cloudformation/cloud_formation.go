package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

type Config struct {
	Client CF
}

type CloudFormation struct {
	client CF
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

// DescribeOutputsAndStatus returns stack outputs, stack status and error. The
// stack status is returned when the error is nil or the error is matched by
// IsOutputsNotAccessible.
func (c *CloudFormation) DescribeOutputsAndStatus(stackName string) ([]Output, string, error) {
	// At first we fetch the CF stack state by describing it via the AWS golang
	// SDK. We are interested in the stack outputs and the stack status, since
	// this tells us if we are able to access outputs at all.
	var stackOutputs []Output
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

		if len(o.Stacks) != 1 {
			return nil, "", microerror.Maskf(tooManyStacksError, "expected 1 stack, got %d", len(o.Stacks))
		}

		stackOutputs = ToOutputs(o.Stacks[0].Outputs)
		stackStatus = *o.Stacks[0].StackStatus
	}

	// We call DescribeOutputsAndStatus in certain GetCurrentState
	// implementations. If the stack creation failed, the outputs can be
	// unaccessible. This can lead to a stack that cannot be deleted. it can also
	// be called during creation, while the outputs are still not accessible.
	var outputsAccessible bool
	{
		okStatuses := []string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusUpdateComplete,
			cloudformation.StackStatusUpdateRollbackComplete,
		}

		for _, s := range okStatuses {
			if stackStatus == s {
				outputsAccessible = true
				break
			}
		}
	}

	if !outputsAccessible {
		return nil, stackStatus, microerror.Maskf(outputsNotAccessibleError, "stack state '%s'", stackStatus)
	}

	return stackOutputs, stackStatus, nil
}

func GetOutputValue(outputs []Output, key string) (string, error) {
	for _, o := range outputs {
		if o.OutputKey == key {
			return o.OutputValue, nil
		}
	}

	return "", microerror.Maskf(outputNotFoundError, "stack output value for key '%s'", key)
}

func ToOutputs(outputs []*cloudformation.Output) []Output {
	var newOutputs []Output

	for _, o := range outputs {
		n := Output{
			OutputKey:   *o.OutputKey,
			OutputValue: *o.OutputValue,
		}

		newOutputs = append(newOutputs, n)
	}

	return newOutputs
}
