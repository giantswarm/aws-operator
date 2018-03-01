package lifecycle

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/giantswarm/aws-operator/service/awsconfig/v7/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/resourcecanceledcontext"
)

// TODO this is heavily copied from the cloudformation resource. It would be
// good to consolidate such common functionality. The common functionality here
// is to fetch output values from the CF stack.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// In order to compute the current state of the guest cluster's main stack we
	// have to describe the CF stacks and lookup the right stack. We dispatch our
	// custom StackState structure and enrich it with all information necessary to
	// reconcile the cloudformation resource.
	stackName := key.MainGuestStackName(customObject)
	describeOutput := &cloudformation.DescribeStacksOutput{}
	{
		r.logger.LogCtx(ctx, "debug", "looking for main stack")

		describeInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}
		describeOutput, err = r.clients.CloudFormation.DescribeStacks(describeInput)
		if IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find main stack")
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
		if len(describeOutput.Stacks) > 1 {
			return nil, microerror.Mask(notFoundError)
		}

		r.logger.LogCtx(ctx, "debug", "found main stack")
	}

	// GetCurrentState is called on cluster deletion, if the stack creation failed
	// the outputs can be unaccessible, this can lead to a stack that cannot be
	// deleted. it can also be called during creation, while the outputs are still
	// not accessible.
	status := *describeOutput.Stacks[0].StackStatus
	{
		errorStatuses := []string{
			"ROLLBACK_IN_PROGRESS",
			"ROLLBACK_COMPLETE",
			"CREATE_IN_PROGRESS",
		}

		for _, errorStatus := range errorStatuses {
			if status == errorStatus {
				return nil, nil
			}
		}
	}

	// In case the current guest cluster is already being updated, we cancel the
	// reconciliation until the current update is done in order to reduce
	// unnecessary friction.
	if status == cloudformation.ResourceStatusUpdateInProgress {
		r.logger.LogCtx(ctx, "debug", fmt.Sprintf("main stack is in state '%s'", cloudformation.ResourceStatusUpdateInProgress))
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "debug", "canceling resource for custom object")

		return nil, nil
	}

	// Here we lookup the worker ASG name.
	var workerASGName string
	{
		workerASGName, err = getStackOutputValue(describeOutput.Stacks[0].Outputs, key.WorkerASGNameOutputKey)
		if IsNotFound(err) {
			// Since we are transitioning between versions we will have situations in
			// which old clusters are updated to new versions and miss the worker ASG
			// name in the CF stack outputs. We cancel the reconciliation until the
			// current update is done in order to reduce unnecessary friction.
			//
			// TODO remove this condition as soon as all guest clusters in existence
			// obtain a worker ASG name.
			r.logger.LogCtx(ctx, "debug", "no worker ASG name")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "debug", "canceling resource for custom object")

			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var lifecycleHooks []*autoscaling.LifecycleHook
	{
		i := &autoscaling.DescribeLifecycleHooksInput{
			AutoScalingGroupName: aws.String(workerASGName),
		}

		o, err := r.clients.AutoScaling.DescribeLifecycleHooks(i)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				r.logger.LogCtx(ctx, "code", aerr.Code(), "level", "error", "message", "describing lifecycle hooks", "stack", fmt.Sprintf("%#v\n", err))
			} else {
				r.logger.LogCtx(ctx, "level", "error", "message", "describing lifecycle hooks", "stack", fmt.Sprintf("%#v\n", err))
			}
		}

		lifecycleHooks = o.LifecycleHooks
	}

	if len(lifecycleHooks) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no lifecycle hooks found")
		return nil, nil
	}

	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	for _, l := range lifecycleHooks {
		fmt.Printf("l.GoString(): %s\n", l.GoString())
		fmt.Printf("\n")
		fmt.Printf("l.String(): %s\n", l.String())
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")

	return nil, nil
}
