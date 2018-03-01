package lifecycle

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/giantswarm/aws-operator/service/awsconfig/v7/key"
)

const (
	WorkerASGName = "WorkerAutoScalingGroup"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	i := &autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: aws.String(key.WorkerASGName),
	}
	o, err := r.clients.AutoScaling.DescribeLifecycleHooks(i)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			r.logger.LogCtx(ctx, "code", aerr.Code(), "level", "error", "message", "describing lifecycle hooks", "stack", fmt.Sprintf("%#v\n", err))
		} else {
			r.logger.LogCtx(ctx, "level", "error", "message", "describing lifecycle hooks", "stack", fmt.Sprintf("%#v\n", err))
		}
	}

	if len(o.LifecycleHooks) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no lifecycle hooks found")
		return nil, nil
	}

	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	for _, l := range o.LifecycleHooks {
		fmt.Printf("l.GoString(): %s\n", l.GoString())
		fmt.Printf("\n")
		fmt.Printf("l.String(): %s\n", l.String())
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	return nil, nil
}
