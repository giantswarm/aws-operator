package aws

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

	"github.com/giantswarm/e2esetup/aws/env"
)

const (
	awsOperatorArnKey     = "aws.awsoperator.arn"
	awsResourceValuesFile = "/tmp/aws-operator-values.yaml"
	credentialName        = "credential-default"
	credentialNamespace   = "giantswarm"
	provider              = "aws"
)

func Setup(ctx context.Context, m *testing.M, config Config) error {
	var v int
	var err error
	var errors []error

	if config.AWSClient == nil {
		return microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.Guest == nil {
		return microerror.Maskf(invalidConfigError, "%T.Guest must not be empty", config)
	}
	if config.Host == nil {
		return microerror.Maskf(invalidConfigError, "%T.Host must not be empty", config)
	}

	vpcPeerID, err := installHostPeerVPC(config)
	if err != nil {
		errors = append(errors, err)
		v = 1
	}

	err = config.Host.Setup()
	if err != nil {
		errors = append(errors, err)
		v = 1
	}

	err = installResources(config, vpcPeerID)
	if err != nil {
		errors = append(errors, err)
		v = 1
	}

	if v == 0 {
		err = config.Guest.Setup()
		if err != nil {
			errors = append(errors, err)
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if env.KeepResources() != "true" {
		config.Host.DeleteGuestCluster(ctx, provider)

		// only do full teardown when not on CI
		if env.CircleCI() != "true" {
			err := teardown(config)
			if err != nil {
				errors = append(errors, err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			config.Host.Teardown()
		}
	}

	if len(errors) > 0 {
		return microerror.Mask(errors[0])
	}

	return nil
}

func installAWSOperator(config Config) error {
	var err error

	var values string
	{
		c := chartvalues.AWSOperatorConfig{
			Provider: chartvalues.AWSOperatorConfigProvider{
				AWS: chartvalues.AWSOperatorConfigProviderAWS{
					Encrypter: "kms",
					Region:    env.AWSRegion(),
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

func installAWSConfig(config Config, vpcPeerID string) error {
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
				VPCPeerID:         vpcPeerID,
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

func installHostPeerVPC(config Config) (string, error) {
	log.Printf("creating Host Peer VPC stack")

	os.Setenv("AWS_ROUTE_TABLE_0", env.AWSRouteTable0())
	os.Setenv("AWS_ROUTE_TABLE_1", env.AWSRouteTable1())
	os.Setenv("CLUSTER_NAME", env.ClusterID())
	stackName := "host-peer-" + env.ClusterID()
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(os.ExpandEnv(e2etemplates.AWSHostVPCStack)),
		TimeoutInMinutes: aws.Int64(2),
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
	var vpcPeerID string
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCID" {
			vpcPeerID = *o.OutputValue
			break
		}
	}
	log.Printf("created Host Peer VPC stack")
	return vpcPeerID, nil
}

func installResources(config Config, vpcPeerID string) error {
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
		err = installCredential(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = installAWSConfig(config, vpcPeerID)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
