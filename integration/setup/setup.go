// +build k8srequired

package setup

import (
	"context"
	"log"
	"os"
	"testing"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	err = config.Host.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = installResources(ctx, config)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = CreateTenantCluster(ctx, config)
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
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installResources(ctx context.Context, config Config) error {
	var err error

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
		err = installCredential(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
