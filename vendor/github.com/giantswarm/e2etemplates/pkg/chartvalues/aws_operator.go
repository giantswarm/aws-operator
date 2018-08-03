package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

const (
	defaultEncrypter = "kms"
)

type AWSOperatorConfig struct {
	Guest    AWSOperatorConfigGuest
	Provider AWSOperatorConfigProvider
	Secret   AWSOperatorConfigSecret

	RegistryPullSecret string
}

type AWSOperatorConfigGuest struct {
	Update AWSOperatorConfigGuestUpdate
}

type AWSOperatorConfigGuestUpdate struct {
	Enabled bool
}

type AWSOperatorConfigProvider struct {
	AWS AWSOperatorConfigProviderAWS
}

type AWSOperatorConfigProviderAWS struct {
	Encrypter string
	Region    string
}

type AWSOperatorConfigSecret struct {
	AWSOperator AWSOperatorConfigSecretAWSOperator
}

type AWSOperatorConfigSecretAWSOperator struct {
	IDRSAPub   string
	SecretYaml AWSOperatorConfigSecretAWSOperatorSecretYaml
}

type AWSOperatorConfigSecretAWSOperatorSecretYaml struct {
	Service AWSOperatorConfigSecretAWSOperatorSecretYamlService
}

type AWSOperatorConfigSecretAWSOperatorSecretYamlService struct {
	AWS AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWS
}

type AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWS struct {
	AccessKey     AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey
	HostAccessKey AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey
}

type AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey struct {
	ID     string
	Secret string
	Token  string
}

// NewAWSOperator renders values required by aws-operator-chart.
func NewAWSOperator(config AWSOperatorConfig) (string, error) {
	if config.Provider.AWS.Region == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Provider.AWS.Region must not be empty", config)
	}
	if config.Secret.AWSOperator.IDRSAPub == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.IDRSAPub must not be empty", config)
	}
	if config.Provider.AWS.Encrypter == "" {
		config.Provider.AWS.Encrypter = defaultEncrypter
	}
	if config.Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.ID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.ID must not be empty", config)
	}
	if config.Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.Secret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.Secret must not be empty", config)
	}
	if config.Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.Token == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.Token must not be empty", config)
	}
	if config.Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.ID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.ID must not be empty", config)
	}
	if config.Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.Secret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.Secret must not be empty", config)
	}
	if config.Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.Token == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.Token must not be empty", config)
	}
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}

	values, err := render.Render(awsOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
