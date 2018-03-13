// +build k8srequired

package integration

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/integration/template"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
)

const (
	awsResourceValuesFile = "/tmp/aws-operator-values.yaml"
)

func installAWSResource() error {
	awsResourceChartValuesEnv := os.ExpandEnv(template.AWSResourceChartValues)
	d := []byte(awsResourceChartValuesEnv)

	err := ioutil.WriteFile(awsResourceValuesFile, d, 0644)
	if err != nil {
		return microerror.Mask(err)
	}

	err = framework.HelmCmd("registry install quay.io/giantswarm/aws-resource-lab-chart:stable -- -n aws-resource-lab --values " + awsResourceValuesFile)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func setup() error {
	var err error

	{
		// TODO configure chart values like for the other operators below.
		err = f.InstallCertOperator()
		if err != nil {
			return microerror.Mask(err)
		}
		err = f.InstallNodeOperator(template.NodeOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = f.InstallAwsOperator(template.AWSOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = f.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}
		// TODO this should probably be in the e2e-harness framework as well just like
		// the other stuff.
		err = installAWSResource()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		logEntry := "created the guest cluster main stack"
		operatorPodName, err := f.PodName("giantswarm", "app=aws-operator")
		if err != nil {
			return microerror.Mask(err)
		}
		err = f.WaitForPodLog("giantswarm", logEntry, operatorPodName)
		if err != nil {
			return microerror.Mask(err)
		}
		err = f.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func setupHostPeerVPC() error {
	log.Printf("Creating Host Peer VPC stack")

	os.Setenv("AWS_ROUTE_TABLE_0", ClusterID()+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", ClusterID()+"_1")
	stackName := "host-peer-" + ClusterID()
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(os.ExpandEnv(template.AWSHostVPCStack)),
		TimeoutInMinutes: aws.Int64(2),
	}
	_, err := c.CloudFormation.CreateStack(stackInput)
	if err != nil {
		return microerror.Mask(err)
	}
	err = c.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	describeOutput, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCID" {
			os.Setenv("AWS_VPC_PEER_ID", *o.OutputValue)
			break
		}
	}
	log.Printf("Host Peer VPC stack created")
	return nil
}
