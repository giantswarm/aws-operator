// +build k8srequired

package setup

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/client"
	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/teardown"
	"github.com/giantswarm/aws-operator/integration/template"
)

const (
	awsResourceValuesFile = "/tmp/aws-operator-values.yaml"
)

func HostPeerVPC(c *client.AWS, g *framework.Guest, h *framework.Host) error {
	log.Printf("Creating Host Peer VPC stack")

	os.Setenv("AWS_ROUTE_TABLE_0", env.ClusterID()+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", env.ClusterID()+"_1")
	stackName := "host-peer-" + env.ClusterID()
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

func Resources(c *client.AWS, g *framework.Guest, h *framework.Host) error {
	var err error

	{
		// TODO configure chart values like for the other operators below.
		err = h.InstallStableOperator("cert-operator", "certconfig", template.CertOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = h.InstallStableOperator("node-operator", "nodeconfig", template.NodeOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = h.InstallBranchOperator("aws-operator", "awsconfig", template.AWSOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = h.InstallCertResource()
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

	return nil
}

func WrapTestMain(c *client.AWS, g *framework.Guest, h *framework.Host, m *testing.M) {
	var v int
	var err error

	err = HostPeerVPC(c, g, h)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = h.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = Resources(c, g, h)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = g.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		h.DeleteGuestCluster()

		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown.Teardown(c, g, h)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			h.Teardown()
		}

		err := teardown.HostPeerVPC(c, g, h)
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
	}

	os.Exit(v)
}

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
