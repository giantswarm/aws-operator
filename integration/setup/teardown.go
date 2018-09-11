// +build k8srequired

package setup

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
)

func teardown(config Config) error {
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
			err = framework.HelmCmd(fmt.Sprintf("delete %s --purge", release))
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete release %#q", release), "stack", fmt.Sprintf("%#v", err))
				errors = append(errors, microerror.Mask(err))
			}
		}
	}

	{
		err = deleteHostPeerVPC(config)
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete host peering VPC"), "stack", fmt.Sprintf("%#v", err))
			errors = append(errors, microerror.Mask(err))
		}
	}

	if len(errors) > 0 {
		return microerror.Mask(errors[0])
	}

	return nil
}

func deleteHostPeerVPC(config Config) error {
	log.Printf("Deleting Host Peer VPC stack")

	_, err := config.AWSClient.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String("host-peer-" + env.ClusterID()),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
