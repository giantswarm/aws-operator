// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
)

func teardown(ctx context.Context, clusterID string, config Config) error {
	var err error
	var errors []error

	{
		releases := []string{
			fmt.Sprintf("%s-aws-operator", config.Host.TargetNamespace()),
			fmt.Sprintf("%s-cert-operator", config.Host.TargetNamespace()),
			fmt.Sprintf("%s-node-operator", config.Host.TargetNamespace()),

			fmt.Sprintf("e2esetup-awsconfig-%s", env.ClusterID()),
			fmt.Sprintf("e2esetup-certs-%s", env.ClusterID()),
		}

		for _, release := range releases {
			err = config.Release.EnsureDeleted(ctx, release)
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete release %#q", release), "stack", fmt.Sprintf("%#v", err))
				errors = append(errors, microerror.Mask(err))
			}
		}
	}

	{
		err = ensureBastionHostDeleted(ctx, clusterID, config)
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", "failed to delete bastion host", "stack", fmt.Sprintf("%#v", err))
			errors = append(errors, microerror.Mask(err))
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

func ensureBastionHostDeleted(ctx context.Context, clusterID string, config Config) error {
	var err error

	var instanceID string
	{
		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("giantswarm.io/cluster"),
					Values: []*string{
						aws.String(clusterID),
					},
				},
				{
					Name: aws.String("giantswarm.io/instance"),
					Values: []*string{
						aws.String("e2e-bastion"),
					},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			return microerror.Maskf(notExistsError, "master instance")
		}
		if len(o.Reservations) != 1 {
			return microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations))
		}
		if len(o.Reservations[0].Instances) != 1 {
			return microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations[0].Instances))
		}

		instanceID = *o.Reservations[0].Instances[0].InstanceId
	}

	{
		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		_, err = config.AWSClient.EC2.TerminateInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
