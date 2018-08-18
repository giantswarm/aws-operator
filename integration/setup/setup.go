// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
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

func Setup(ctx context.Context, m *testing.M, config Config) {
	err := config.Validate()
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "error during e2e config validation", "stack", fmt.Sprintf("%#v", err))
		os.Exit(1)
	}

	exitCode, err := setup(m, config)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "error during e2e setup", "stack", fmt.Sprintf("%#v", err))
		os.Exit(1)
	}

	os.Exit(exitCode)
}

func setup(m *testing.M, config Config) (int, error) {
	ctx := context.Background()

	vpcPeerID, err := installHostPeerVPC(ctx, config)
	if err != nil {
		return 0, microerror.Mask(err)
	} else {
		defer teardownHostPeerVPC(ctx, config)
	}

	err = config.Host.Setup()
	if err != nil {
		return 0, microerror.Mask(err)
	} else if !env.KeepResources() && !env.CircleCI() {
		config.Host.Teardown()
	}

	err = installResources(ctx, config, vpcPeerID)
	if err != nil {
		return 0, microerror.Mask(err)
	} else if !env.KeepResources() && !env.CircleCI() {
		defer teardownResources(ctx, config)
	}

	err = config.Guest.Setup()
	if err != nil {
		return 0, microerror.Mask(err)
	}

	code := m.Run()

	if !env.KeepResources() {
		err := config.Host.DeleteGuestCluster(ctx, provider)
		if err != nil {
			return 0, microerror.Mask(err)
		}
	}

	return code, nil
}

func installAWSOperator(ctx context.Context, config Config) error {
	config.Logger.LogCtx(ctx, "level", "debug", "message", "installing aws-operator")

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
			config.Logger.LogCtx(ctx, "level", "debug", "message", "did not install aws-operator")
			return microerror.Mask(err)
		}

	}

	err = config.Host.InstallBranchOperator("aws-operator", "awsconfig", values)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not install aws-operator")
		return microerror.Mask(err)
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", "installed aws-operator")
	return nil
}

func installAWSConfig(ctx context.Context, config Config, vpcPeerID string) error {
	config.Logger.LogCtx(ctx, "level", "debug", "message", "installing awsconfig")

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
			config.Logger.LogCtx(ctx, "level", "debug", "message", "did not install awsconfig")
			return microerror.Mask(err)
		}
	}

	err = config.Host.InstallResource("apiextensions-aws-config-e2e", values, ":stable")
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not install awsconfig")
		return microerror.Mask(err)
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", "installed awsconfig")
	return nil
}

func installCredential(ctx context.Context, config Config) error {
	config.Logger.LogCtx(ctx, "level", "debug", "message", "installing credential secret")
	var err error

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
	n := backoff.NewNotifier(config.Logger, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not install credential secret")
		return microerror.Mask(err)
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", "installed credential secret")
	return nil
}

func installHostPeerVPC(ctx context.Context, config Config) (string, error) {
	config.Logger.LogCtx(ctx, "level", "debug", "message", "creating host peer VPC stack")

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
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not create host peer VPC stack")
		return "", microerror.Mask(err)
	}
	err = config.AWSClient.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not create host peer VPC stack")
		return "", microerror.Mask(err)
	}
	describeOutput, err := config.AWSClient.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "debug", "message", "did not create host peer VPC stack")
		return "", microerror.Mask(err)
	}
	var vpcPeerID string
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCID" {
			vpcPeerID = *o.OutputValue
			break
		}
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", "created host peer VPC stack")
	return vpcPeerID, nil
}

func installResources(ctx context.Context, config Config, vpcPeerID string) error {
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
		err = installAWSOperator(ctx, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = config.Host.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}
		err = installCredential(ctx, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = installAWSConfig(ctx, config, vpcPeerID)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
