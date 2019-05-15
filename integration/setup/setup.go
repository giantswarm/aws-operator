// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/key"
)

const (
	provider = "aws"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	var v int
	var err error

	err = installResources(ctx, config)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "failed to install AWS resources", "stack", fmt.Sprintf("%#v", err))
		v = 1
	}

	{
		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			o := func() error {
				//err = ensureBastionHostCreated(ctx, env.ClusterID(), config)
				//if err != nil {
				//	return microerror.Mask(err)
				//}

				return nil
			}
			b := backoff.NewMaxRetries(10, 1*time.Minute)
			n := backoff.NewNotifier(config.Logger, ctx)

			err := backoff.RetryNotify(o, b, n)
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "error", "message", err.Error())
			}

			return nil
		})

		g.Go(func() error {
			if v == 0 && config.UseDefaultTenant {
				wait := true
				err = EnsureTenantClusterCreated(ctx, env.ClusterID(), config, wait)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			return nil
		})

		err := g.Wait()
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", err.Error())
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if !env.KeepResources() {
		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			o := func() error {
				//err = ensureBastionHostDeleted(ctx, env.ClusterID(), config)
				//if err != nil {
				//	return microerror.Mask(err)
				//}

				return nil
			}
			b := backoff.NewMaxRetries(10, 1*time.Minute)
			n := backoff.NewNotifier(config.Logger, ctx)

			err := backoff.RetryNotify(o, b, n)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})

		g.Go(func() error {
			if config.UseDefaultTenant {
				wait := true
				err := EnsureTenantClusterDeleted(ctx, env.ClusterID(), config, wait)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			return nil
		})

		err := g.Wait()
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", err.Error())
			v = 1
		}

		if !env.CircleCI() {
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
	vpcID, err := ensureHostVPCCreated(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	var err error

	var values string
	{
		c := chartvalues.AWSOperatorConfig{
			Provider: chartvalues.AWSOperatorConfigProvider{
				AWS: chartvalues.AWSOperatorConfigProviderAWS{
					Encrypter:       "kms",
					Region:          env.AWSRegion(),
					RouteTableNames: env.AWSRouteTable0() + "," + env.AWSRouteTable1(),
				},
			},
			Secret: chartvalues.AWSOperatorConfigSecret{
				AWSOperator: chartvalues.AWSOperatorConfigSecretAWSOperator{
					CredentialDefault: chartvalues.AWSOperatorConfigSecretAWSOperatorCredentialDefault{
						AdminARN:       env.GuestAWSARN(),
						AWSOperatorARN: env.GuestAWSARN(),
					},
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

	err = config.Release.InstallOperator(ctx, key.AWSOperatorReleaseName(), release.NewVersion(env.CircleSHA()), values, providerv1alpha1.NewAWSConfigCRD())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installResources(ctx context.Context, config Config) error {
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

		err = config.Release.Install(ctx, key.VaultReleaseName(), release.NewStableVersion(), values, config.Release.Condition().PodExists(ctx, "default", "app=vault"))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.CertOperatorConfig{
			CommonDomain:       env.CommonDomain(),
			RegistryPullSecret: env.RegistryPullSecret(),
			Vault: chartvalues.CertOperatorVault{
				Token: env.VaultToken(),
			},
		}

		values, err := chartvalues.NewCertOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.InstallOperator(ctx, key.CertOperatorReleaseName(), release.NewStableVersion(), values, corev1alpha1.NewCertConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c := chartvalues.NodeOperatorConfig{
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err := chartvalues.NewNodeOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Release.InstallOperator(ctx, key.NodeOperatorReleaseName(), release.NewStableVersion(), values, corev1alpha1.NewDrainerConfigCRD())
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

	return nil
}

func ensureHostVPCCreated(ctx context.Context, config Config) (string, error) {
	stackName := "host-peer-" + env.ClusterID()

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating stack %#q", stackName))

		os.Setenv("AWS_ROUTE_TABLE_0", env.AWSRouteTable0())
		os.Setenv("AWS_ROUTE_TABLE_1", env.AWSRouteTable1())
		os.Setenv("CLUSTER_NAME", env.ClusterID())
		stackInput := &cloudformation.CreateStackInput{
			StackName:        aws.String(stackName),
			TemplateBody:     aws.String(os.ExpandEnv(e2etemplates.AWSHostVPCStack)),
			TimeoutInMinutes: aws.Int64(2),
		}
		_, err := config.AWSClient.CloudFormation.CreateStack(stackInput)
		if IsStackAlreadyExists(err) {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("stack %#q is already created", stackName))
		} else if err != nil {
			return "", microerror.Mask(err)
		} else {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created stack %#q", stackName))
		}
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for stack %#q complete status", stackName))

		err := config.AWSClient.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return "", microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for stack %#q complete status", stackName))
	}

	var vpcID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding `VPCID` output in stack %#q", stackName))

		describeOutput, err := config.AWSClient.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return "", microerror.Mask(err)
		}
		for _, o := range describeOutput.Stacks[0].Outputs {
			if *o.OutputKey == "VPCID" {
				vpcID = *o.OutputValue
				break
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found `VPCID` output in stack %#q", stackName))
	}

	return vpcID, nil
}
