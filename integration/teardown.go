// +build k8srequired

package integration

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
)

func teardownHostPeerVPC() error {
	log.Printf("Deleting Host Peer VPC stack")

	_, err := c.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String("host-peer-" + ClusterID()),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func teardown() error {
	var err error

	{
		err = framework.HelmCmd("delete aws-operator --purge")
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd("delete cert-operator --purge")
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd("delete node-operator --purge")
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = framework.HelmCmd("delete cert-resource-lab --purge")
		if err != nil {
			return microerror.Mask(err)
		}
		err = framework.HelmCmd("delete aws-resource-lab --purge")
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
