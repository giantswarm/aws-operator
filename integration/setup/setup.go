// +build k8srequired

package setup

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
)

const (
	awsResourceValuesFile = "/tmp/aws-operator-values.yaml"
	provider              = "aws"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	// Perform teardown before execution.
	{
		err := teardown(ctx, config)
		if err != nil {
			os.Exit(1)
		}
	}

	var v int
	var err error

	vpcPeerID, err := installHostPeerVPC(config)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = installResources(ctx, config, vpcPeerID)
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
			err := teardown(ctx, config)
			if err != nil {
				// teardown errors are logged inside the function.
				v = 1
			}
		}
	}

	os.Exit(v)
}

func installAWSOperator(ctx context.Context, config Config) error {
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
					CredentialDefault: chartvalues.AWSOperatorConfigSecretAWSOperatorCredentialDefault{
						AdminARN:       env.GuestAWSARN(),
						AWSOperatorARN: env.GuestAWSARN(),
					},
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

	err = config.Release.InstallOperator(ctx, "aws-operator", release.NewVersion(env.CircleSHA()), values, providerv1alpha1.NewAWSConfigCRD())
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func installAWSConfig(ctx context.Context, config Config, vpcPeerID string) error {
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

	err = config.Release.Install(ctx, "apiextensions-aws-config-e2e", release.NewStableVersion(), values)
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

func installResources(ctx context.Context, config Config, vpcPeerID string) error {
	var err error

	{
		err := config.K8s.EnsureNamespaceCreated(ctx, namespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.E2ESetupVaultConfig{
			Vault: chartvalues.E2ESetupVaultConfigVault{
				Token: env.VaultToken(),
			},
		}

		values, err := chartvalues.NewE2ESetupVault(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.Install(ctx, "e2esetup-vault", release.NewStableVersion(), values, config.Release.Condition().PodExists(ctx, "default", "app=vault"))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		var values string
		{
			c := chartvalues.CertOperatorConfig{
				ClusterName:        env.ClusterID(),
				CommonDomain:       env.CommonDomain(),
				RegistryPullSecret: env.RegistryPullSecret(),
				Vault: chartvalues.CertOperatorVault{
					Token: env.VaultToken(),
				},
			}

			values, err = chartvalues.NewCertOperator(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		err = config.Release.InstallOperator(ctx, "cert-operator", release.NewStableVersion(), values, corev1alpha1.NewCertConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		var values string
		{
			c := chartvalues.NodeOperatorConfig{
				RegistryPullSecret: env.RegistryPullSecret(),
			}

			values, err = chartvalues.NewNodeOperator(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		err = config.Release.InstallOperator(ctx, "node-operator", release.NewStableVersion(), values, corev1alpha1.NewNodeConfigCRD())
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
	}

	{
		err = installAWSConfig(ctx, config, vpcPeerID)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
