// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
)

func teardown(ctx context.Context, config Config) error {
	var err error
	var errors []error

	{
		releases := []string{
			fmt.Sprintf("%s-aws-operator", config.Host.TargetNamespace()),
			fmt.Sprintf("%s-cert-operator", config.Host.TargetNamespace()),
			fmt.Sprintf("%s-node-operator", config.Host.TargetNamespace()),

			fmt.Sprintf("%s-cert-config-e2e", config.Host.TargetNamespace()),
			fmt.Sprintf("%s-aws-config-e2e", config.Host.TargetNamespace()),
		}

		for _, release := range releases {
			err = config.Resource.EnsureDeleted(ctx, release)
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete release %#q", release), "stack", fmt.Sprintf("%#v", err))
				errors = append(errors, microerror.Mask(err))
			}
		}
	}

	{
		stackName := "host-peer-" + env.ClusterID()
		wait := true

		// TODO extract cloudformation stack client which can be also
		// used in cloudformation resource.
		// Issue: https://github.com/giantswarm/giantswarm/issues/3783.
		err := ensureDeletedStack(ctx, config, stackName, wait)
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete host VPC stack"), "stack", fmt.Sprintf("%#v", err))
			errors = append(errors, microerror.Mask(err))
		}
	}

	{
		// TODO there should be error handling for the framework teardown.
		config.Host.Teardown()
	}

	if len(errors) > 0 {
		return microerror.Mask(errors[0])
	}

	return nil
}

func ensureDeletedStack(ctx context.Context, config Config, stackName string, wait bool) error {
	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring deletion of stack %#q", stackName))

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out stack %#q", stackName))

		_, err := getStack(ctx, config, stackName)
		if IsNotFound(err) {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find stack %#q", stackName))
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found stack %#q", stackName))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("triggering stack %#q deletion", stackName))

		err := startStackDeletion(ctx, config, stackName)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("triggered stack %#q deletion", stackName))
	}

	if wait {
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for stack %#q to be deleted", stackName))

		o := func() error {
			_, err := getStack(ctx, config, stackName)
			if IsNotFound(err) {
				return nil
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}

			return microerror.Mask(stillExistsError)
		}
		b := backoff.NewConstant(5*time.Minute, 20*time.Second)
		n := backoff.NewNotifier(config.Logger, context.Background())
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for stack %#q to be deleted", stackName))
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured deletion of stack %#q", stackName))
	return nil
}

func getStack(ctx context.Context, config Config, stackName string) (*cloudformation.Stack, error) {
	in := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	out, err := config.AWSClient.CloudFormation.DescribeStacks(in)
	if IsStackNotFound(err) {
		return nil, microerror.Maskf(notFoundError, err.Error())
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(out.Stacks) == 0 {
		return nil, microerror.Mask(notFoundError)
	}
	if len(out.Stacks) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return out.Stacks[0], nil
}

func startStackDeletion(ctx context.Context, config Config, stackName string) error {
	in := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	}

	_, err := config.AWSClient.CloudFormation.DeleteStack(in)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
