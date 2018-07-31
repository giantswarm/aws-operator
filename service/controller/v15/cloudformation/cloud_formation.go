package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

type Config struct {
	Client CloudFormationInterface
}

type CloudFormation struct {
	client CloudFormationInterface
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

func (c *CloudFormation) DescribeOutputsAndStatus(stackName string) ([]*cloudformation.Output, error) {
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
			return nil, microerror.Maskf(stackNotFoundError, "stack name '%s'", stackName)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(o.Stacks) > 1 {
			return nil, microerror.Maskf(tooManyStacksError, "expected 1 stack, got %d", len(o.Stacks))
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
			cloudformation.StackStatusCreateInProgress,
			cloudformation.StackStatusRollbackInProgress,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusUpdateInProgress,
		}

		// TODO to discuss in the PR:
		//
		//
		// Complete states:
		//
		//	StackStatusDeleteComplete = "DELETE_COMPLETE"
		//	StackStatusUpdateComplete = "UPDATE_COMPLETE"
		//	StackStatusCreateComplete = "CREATE_COMPLETE"
		//
		//
		// Rollback complete states:
		//
		//	StackStatusRollbackComplete = "ROLLBACK_COMPLETE"
		//	StackStatusUpdateRollbackComplete = "UPDATE_ROLLBACK_COMPLETE"
		//
		//
		// Failed states:
		//
		//	StackStatusCreateFailed = "CREATE_FAILED"
		//	StackStatusRollbackFailed = "ROLLBACK_FAILED"
		//	StackStatusDeleteFailed = "DELETE_FAILED"
		//	StackStatusUpdateRollbackFailed = "UPDATE_ROLLBACK_FAILED"
		//
		//
		// In progress states:
		//
		//	StackStatusCreateInProgress = "CREATE_IN_PROGRESS"
		//	StackStatusDeleteInProgress = "DELETE_IN_PROGRESS"
		//	StackStatusReviewInProgress = "REVIEW_IN_PROGRESS"
		//	StackStatusRollbackInProgress = "ROLLBACK_IN_PROGRESS"
		//	StackStatusUpdateCompleteCleanupInProgress = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"
		//	StackStatusUpdateInProgress = "UPDATE_IN_PROGRESS"
		//	StackStatusUpdateRollbackCompleteCleanupInProgress = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"
		//	StackStatusUpdateRollbackInProgress = "UPDATE_ROLLBACK_IN_PROGRESS"

		for _, s := range errorStatuses {
			if stackStatus == s {
				return nil, microerror.Maskf(stackInTransitionError, "due to stack state '%s'", stackStatus)
			}
		}
	}

	return stackOutputs, nil
}

func (c *CloudFormation) GetOutputValue(outputs []*cloudformation.Output, key string) (string, error) {
	for _, o := range outputs {
		if *o.OutputKey == key {
			return *o.OutputValue, nil
		}
	}

	return "", microerror.Maskf(outputNotFoundError, "stack output value for key '%s'", key)
}
