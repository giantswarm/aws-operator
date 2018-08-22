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
		err = framework.HelmCmd(fmt.Sprintf("delete %s-aws-operator --purge", config.Host.TargetNamespace()))
		if err != nil {
			errors = append(errors, microerror.Mask(err))
		}
		err = framework.HelmCmd(fmt.Sprintf("delete %s-cert-operator --purge", config.Host.TargetNamespace()))
		if err != nil {
			errors = append(errors, microerror.Mask(err))
		}
		err = framework.HelmCmd(fmt.Sprintf("delete %s-node-operator --purge", config.Host.TargetNamespace()))
		if err != nil {
			errors = append(errors, microerror.Mask(err))
		}
	}

	{
		err = framework.HelmCmd(fmt.Sprintf("delete %s-cert-config-e2e --purge", config.Host.TargetNamespace()))
		if err != nil {
			errors = append(errors, microerror.Mask(err))
		}
		err = framework.HelmCmd(fmt.Sprintf("delete %s-aws-config-e2e --purge", config.Host.TargetNamespace()))
		if err != nil {
			errors = append(errors, microerror.Mask(err))
		}
	}

	if len(errors) > 0 {
		return microerror.Mask(errors[0])
	}

	err = deleteHostPeerVPC(config)
	if err != nil {
		return microerror.Mask(err)
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
