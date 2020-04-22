// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2esetup/privaterepo"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
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

				err = EnsureHostPeerStackDeleted(ctx, config, wait)
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

	os.Exit(v)
}

func installAWSOperator(ctx context.Context, config Config) error {
	sshUserList, err := newSSHUserList(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}
	err = ensureHostPeerStackCreated(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	var values string
	{
		c := chartvalues.AWSOperatorConfig{
			InstallationName: fmt.Sprintf("cp-peer-%s", env.ClusterID()),
			Provider: chartvalues.AWSOperatorConfigProvider{
				AWS: chartvalues.AWSOperatorConfigProviderAWS{
					Encrypter:       "kms",
					Region:          env.AWSRegion(),
					RouteTableNames: env.AWSRouteTable0() + "," + env.AWSRouteTable1(),
				},
			},
			RegistryPullSecret: env.RegistryPullSecret(),
			Secret: chartvalues.AWSOperatorConfigSecret{
				AWSOperator: chartvalues.AWSOperatorConfigSecretAWSOperator{
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
			SSH: chartvalues.AWSOperatorConfigSSH{
				UserList: sshUserList,
			},
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

func installCredentialDefaultSecret(ctx context.Context, config Config) error {
	var err error

	var values string
	{
		c := chartvalues.CredentialdConfig{
			AWS: chartvalues.CredentialdConfigAWS{
				CredentialDefault: chartvalues.CredentialdConfigAWSCredentialDefault{
					AdminARN:       env.GuestAWSARN(),
					AWSOperatorARN: env.GuestAWSARN(),
				},
			},
			Deployment: chartvalues.CredentialdConfigDeployment{
				Replicas: 0,
			},
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err = chartvalues.NewCredentiald(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = config.Release.Install(ctx, key.CredentialdReleaseName(), release.NewStableVersion(), values, config.Release.Condition().SecretExists(ctx, namespace, "credential-default"))
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

	// Install AWSConfig CRD since it is no longer installed by the operator.
	// If the CRD doesn't exist it fails.
	{
		b := backoff.NewMaxRetries(5, 2*time.Second)

		err := config.CPCRDClient.EnsureCreated(ctx, providerv1alpha1.NewAWSConfigCRD(), b)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Install AWSCluster CRD for IPAM resource. It checks clusters objects
	// for allocated ranges. If the CRD doesn't exist it fails.
	{
		b := backoff.NewMaxRetries(5, 2*time.Second)

		err := config.CPCRDClient.EnsureCreated(ctx, infrastructurev1alpha2.NewAWSClusterCRD(), b)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err := config.HelmClient.EnsureTillerInstalled(ctx)
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
		err = installCredentialDefaultSecret(ctx, config)
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

func newSSHUserList(ctx context.Context, config Config) (string, error) {
	var err error

	var privateRepo *privaterepo.PrivateRepo
	{
		c := privaterepo.Config{
			Owner: "giantswarm",
			Repo:  "installations",
			Token: env.GithubToken(),
		}

		privateRepo, err = privaterepo.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var sshUserList string
	{
		content, err := privateRepo.Content(ctx, "default-draughtsman-configmap-values.yaml")
		if err != nil {
			return "", microerror.Mask(err)
		}

		sshUserList, err = privaterepo.ContentToSSHUserList(content)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return sshUserList, nil
}
