// +build k8srequired

package setup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/e2e-harness/pkg/framework"

	"github.com/giantswarm/aws-operator/integration/env"
)

func teardownResources(ctx context.Context, config Config) {
	releases := []string{
		config.Host.TargetNamespace() + "-aws-operator",
		config.Host.TargetNamespace() + "-cert-operator",
		config.Host.TargetNamespace() + "-node-operator",
		config.Host.TargetNamespace() + "-cert-config-e2e",
		config.Host.TargetNamespace() + "-aws-config-e2e",
	}

	for _, release := range releases {
		config.Logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("deleting %s", release))

		err := framework.HelmCmd(fmt.Sprintf("delete %s --purge", release))
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("did not delete %s", release))
			config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("error during %s deletion", release), "stack", fmt.Sprintf("%#v", err))
		}

		config.Logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("deleted %s", release))
	}
}

func teardownHostPeerVPC(ctx context.Context, config Config) {
	config.Logger.LogCtx(ctx, "level", "debug", "deleting host peer VPC stack")

	_, err := config.AWSClient.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String("host-peer-" + env.ClusterID()),
	})
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not delete host VPC stack", "stack", fmt.Sprintf("%#v", err))
		config.Logger.LogCtx(ctx, "level", "error", "message", "error during VPC stack deletion", "stack", fmt.Sprintf("%#v", err))
	}

	config.Logger.LogCtx(ctx, "level", "debug", "deleted host peer VPC stack")
}
