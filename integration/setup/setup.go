// +build k8srequired

package setup

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/integration/env"
)

const (
	awsOperatorArnKey     = "aws.awsoperator.arn"
	awsResourceValuesFile = "/tmp/aws-operator-values.yaml"
	credentialName        = "credential-default"
	credentialNamespace   = "giantswarm"
	provider              = "aws"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	var v int
	var err error

	err = config.Validate()
	if err != nil {
		log.Printf("%#v\n", err)
		os.Exit(1)
	}

	vpcPeerID, encrypterVaultSubnetID, err := installHostPeerVPC(config)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	extConfig := extendedConfig{
		Config:                 config,
		VPCPeerID:              vpcPeerID,
		EncrypterVaultSubnetID: encrypterVaultSubnetID,
	}

	// TODO make encrypter vault install conditional depending on encrypter.

	vaultAddress, err := installEncrypterVault(extConfig)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	extConfig.VaultAddress = vaultAddress

	err = config.Host.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = installResources(extConfig)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		err = config.Guest.Setup()
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		config.Host.DeleteGuestCluster(ctx, provider)

		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown(config)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			config.Host.Teardown()
		}
	}

	os.Exit(v)
}

func installAWSOperator(config extendedConfig) error {
	var err error

	var values string
	{
		c := chartvalues.AWSOperatorConfig{
			Provider: chartvalues.AWSOperatorConfigProvider{
				AWS: chartvalues.AWSOperatorConfigProviderAWS{
					Encrypter:    config.Encrypter,
					Region:       env.AWSRegion(),
					VaultAddress: config.VaultAddress,
				},
			},
			Secret: chartvalues.AWSOperatorConfigSecret{
				AWSOperator: chartvalues.AWSOperatorConfigSecretAWSOperator{
					IDRSAPub: env.IDRSAPub(),
					SecretYaml: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYaml{
						Service: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlService{
							AWS: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWS{
								AccessKey: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey{
									ID:     env.GuestAWSAccessKeyID(),
									Secret: env.GuestAWSAccessKeySecret(),
									Token:  env.GuestAWSAccessKeyToken(),
								},
								HostAccessKey: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey{
									ID:     env.HostAWSAccessKeyID(),
									Secret: env.HostAWSAccessKeySecret(),
									Token:  env.HostAWSAccessKeyToken(),
								},
							},
						},
					},
				},
			},
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err = chartvalues.NewAWSOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

	}

	err = config.Host.InstallBranchOperator("aws-operator", "awsconfig", values)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func installAWSConfig(config extendedConfig) error {
	var err error

	var values string
	{
		c := chartvalues.APIExtensionsAWSConfigE2EConfig{
			CommonDomain:         env.CommonDomain(),
			ClusterName:          env.ClusterID(),
			SSHPublicKey:         env.IDRSAPub(),
			VersionBundleVersion: env.VersionBundleVersion(),

			AWS: chartvalues.APIExtensionsAWSConfigE2EConfigAWS{
				Region:            env.AWSRegion(),
				APIHostedZone:     env.AWSAPIHostedZoneGuest(),
				IngressHostedZone: env.AWSIngressHostedZoneGuest(),
				RouteTable0:       env.AWSRouteTable0(),
				RouteTable1:       env.AWSRouteTable1(),
				VPCPeerID:         config.VPCPeerID,
			},
		}

		values, err = chartvalues.NewAPIExtensionsAWSConfigE2E(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = config.Host.InstallResource("apiextensions-aws-config-e2e", values, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCredential(config Config) error {
	var err error

	var l micrologger.Logger
	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	o := func() error {
		k8sClient := config.Host.K8sClient()

		k8sClient.CoreV1().Secrets(credentialNamespace).Delete(credentialName, &metav1.DeleteOptions{})

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: credentialName,
			},
			Data: map[string][]byte{
				awsOperatorArnKey: []byte(env.GuestAWSARN()),
			},
		}

		_, err := k8sClient.CoreV1().Secrets(credentialNamespace).Create(secret)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installEncrypterVault(config extendedConfig) (string, error) {
	log.Printf("creating encrypter vault")

	stackName := "encrypter-vault-" + env.ClusterID()
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(encrypterVaultTemplate),
		TimeoutInMinutes: aws.Int64(10),
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String("AccessKey"),
				ParameterValue: aws.String(env.HostAWSAccessKeyID()),
			},
			{
				ParameterKey:   aws.String("SecretKeyId"),
				ParameterValue: aws.String(env.HostAWSAccessKeySecret()),
			},
			{
				ParameterKey:   aws.String("SubnetId"),
				ParameterValue: aws.String(config.EncrypterVaultSubnetID),
			},
			{
				ParameterKey:   aws.String("VpcId"),
				ParameterValue: aws.String(config.VPCPeerID),
			},
		},
	}
	_, err := config.AWSClient.CloudFormation.CreateStack(stackInput)
	if err != nil {
		return "", microerror.Mask(err)
	}
	err = config.AWSClient.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", microerror.Mask(err)
	}
	describeOutput, err := config.AWSClient.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", microerror.Mask(err)
	}
	var vaultAddress string
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VaultAddress" {
			vaultAddress = *o.OutputValue
			break
		}
	}
	log.Printf("created encrypter vault")
	return vaultAddress, nil
}

func installHostPeerVPC(config Config) (vpcID, encrypterVaultSubnetID string, err error) {
	log.Printf("creating Host Peer VPC stack")

	os.Setenv("AWS_ROUTE_TABLE_0", env.AWSRouteTable0())
	os.Setenv("AWS_ROUTE_TABLE_1", env.AWSRouteTable1())
	os.Setenv("CLUSTER_NAME", env.ClusterID())
	stackName := "host-peer-" + env.ClusterID()
	stackInput := &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
		// ---
		// ---
		// ---
		// ---
		// TODO Update template in e2etemplates.
		// ---
		// ---
		// ---
		// ---
		TemplateBody:     aws.String(os.ExpandEnv(e2etemplates.AWSHostVPCStack)),
		TimeoutInMinutes: aws.Int64(2),
	}
	_, err = config.AWSClient.CloudFormation.CreateStack(stackInput)
	if err != nil {
		return "", "", microerror.Mask(err)
	}
	err = config.AWSClient.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", "", microerror.Mask(err)
	}
	describeOutput, err := config.AWSClient.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	outputFn := func(key string) (string, error) {
		for _, o := range describeOutput.Stacks[0].Outputs {
			if *o.OutputKey == "VPCID" {
				return *o.OutputValue, nil
			}
		}
		return "", microerror.Maskf(executionFailedError, "CF stack output for key %#q not found", key)
	}

	vpcID, err = outputFn("VPCID")
	if err != nil {
		return "", "", microerror.Mask(err)
	}
	vpcID, err = outputFn("VPCVaultSubnet")
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	log.Printf("created Host Peer VPC stack")
	return vpcID, encrypterVaultSubnetID, nil
}

func installResources(config extendedConfig) error {
	var err error

	{
		// TODO configure chart values like for the other operators below.
		err = config.Host.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
		err = config.Host.InstallStableOperator("node-operator", "drainerconfig", e2etemplates.NodeOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = installAWSOperator(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = config.Host.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}
		err = installCredential(config.Config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = installAWSConfig(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
