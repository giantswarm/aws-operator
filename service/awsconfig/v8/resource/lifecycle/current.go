package lifecycle

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/resourcecanceledcontext"

	cloudformationservice "github.com/giantswarm/aws-operator/service/awsconfig/v8/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v8/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the guest cluster cloud formation stack in the AWS API")

	var stackOutputs []*cloudformation.Output
	{
		stackOutputs, _, err = r.service.DescribeOutputsAndStatus(key.MainGuestStackName(customObject))
		if cloudformationservice.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster cloud formation stack in the AWS API")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil, nil

		} else if cloudformationservice.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster cloud formation stack output values are not accessible due to stack state transition")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster cloud formation stack in the AWS API")

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the guest cluster worker ASG name in the cloud formation stack")

	var workerASGName string
	{
		workerASGName, err = r.service.GetOutputValue(stackOutputs, key.WorkerASGKey)
		if cloudformationservice.IsOutputNotFound(err) {
			// Since we are transitioning between versions we will have situations in
			// which old clusters are updated to new versions and miss the ASG name in
			// the CF stack outputs. We stop resource reconciliation at this point to
			// reconcile again at a later point when the stack got upgraded and
			// contains the ASG name.
			//
			// TODO remove this condition as soon as all guest clusters in existence
			// obtain a ASG name.
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster worker ASG name in the cloud formation stack")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster worker ASG name in the cloud formation stack")

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("looking for the guest cluster instances being in state '%s'", autoscaling.LifecycleStateTerminatingWait))

	var instances []*autoscaling.Instance
	{
		i := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{
				aws.String(workerASGName),
			},
		}

		o, err := r.clients.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, g := range o.AutoScalingGroups {
			for _, i := range g.Instances {
				if *i.LifecycleState == autoscaling.LifecycleStateTerminatingWait {
					instances = append(instances, i)
				}
			}
		}

		if len(instances) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the guest cluster instances being in state '%s'", autoscaling.LifecycleStateTerminatingWait))
			return nil, nil

		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d guest cluster instances being in state '%s'", len(instances), autoscaling.LifecycleStateTerminatingWait))
		}
	}

	// TODO check if NodeConfig exists
	// TODO create NodeConfig if it does not exist
	// TODO check if NodeConfig has final state

	for _, instance := range instances {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completing lifecycle hook action for guest cluster instance '%s'", *instance.InstanceId))

		i := &autoscaling.CompleteLifecycleActionInput{
			AutoScalingGroupName:  aws.String(workerASGName),
			InstanceId:            instance.InstanceId,
			LifecycleActionResult: aws.String("CONTINUE"),
			LifecycleHookName:     aws.String(key.NodeDrainerLifecycleHookName),
		}

		_, err := r.clients.AutoScaling.CompleteLifecycleAction(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("completed lifecycle hook action for guest cluster instance '%s'", *instance.InstanceId))
	}

	return nil, nil
}
