// +build k8srequired

package setup

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/teardown"
)

const (
	awsResourceValuesFile = "/tmp/aws-operator-values.yaml"
	credentialName        = "credential-default"
	awsOperatorArnKey     = "aws.awsoperator.arn"
)

func HostPeerVPC(c *awsclient.Client, g *framework.Guest, h *framework.Host) error {
	log.Printf("Creating Host Peer VPC stack")

	os.Setenv("AWS_ROUTE_TABLE_0", env.ClusterID()+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", env.ClusterID()+"_1")
	stackName := "host-peer-" + env.ClusterID()
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(os.ExpandEnv(e2etemplates.AWSHostVPCStack)),
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

func Resources(c *awsclient.Client, g *framework.Guest, h *framework.Host) error {
	var err error

	{
		// TODO configure chart values like for the other operators below.
		err = h.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = h.InstallStableOperator("node-operator", "nodeconfig", e2etemplates.NodeOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = h.InstallBranchOperator("aws-operator", "awsconfig", e2etemplates.AWSOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = h.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}
		err = installCredential(h)
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

func WrapTestMain(c *awsclient.Client, g *framework.Guest, h *framework.Host, m *testing.M) {
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

	if v == 0 {
		err = g.Setup()
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		name := "aws-operator"
		customResource := "awsconfig"
		logEntry := "deleted the guest cluster main stack"
		h.DeleteGuestCluster(name, customResource, logEntry)

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
	o := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		framework.HelmCmd("delete --purge aws-resource-lab")

		awsResourceChartValuesEnv := os.ExpandEnv(e2etemplates.AWSResourceChartValues)
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
	b := framework.NewExponentialBackoff(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCredential(h *framework.Host) error {
	o := func() error {
		k8sClient := h.K8sClient()

		k8sClient.CoreV1().Secrets("giantswarm").Delete(credentialName, &metav1.DeleteOptions{})

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: credentialName,
			},
			Data: map[string][]byte{
				awsOperatorArnKey: []byte(env.GuestAWSArn()),
			},
		}

		_, err := k8sClient.CoreV1().Secrets("giantswarm").Create(secret)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := framework.NewExponentialBackoff(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
