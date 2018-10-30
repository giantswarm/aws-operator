package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type AzureOperatorConfig struct {
	Provider AzureOperatorConfigProvider
	Secret   AzureOperatorConfigSecret
}

type AzureOperatorConfigProvider struct {
	Azure AzureOperatorConfigProviderAzure
}

type AzureOperatorConfigProviderAzure struct {
	Location string
}

type AzureOperatorConfigSecret struct {
	AzureOperator AzureOperatorConfigSecretAzureOperator
	Registry      AzureOperatorConfigSecretRegistry
}

type AzureOperatorConfigSecretAzureOperator struct {
	CredentialDefault AzureOperatorConfigSecretAzureOperatorCredentialDefault
	SecretYaml        AzureOperatorConfigSecretAzureOperatorSecretYaml
}

type AzureOperatorConfigSecretAzureOperatorCredentialDefault struct {
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
}

type AzureOperatorConfigSecretAzureOperatorSecretYaml struct {
	Service AzureOperatorConfigSecretAzureOperatorSecretYamlService
}

type AzureOperatorConfigSecretAzureOperatorSecretYamlService struct {
	Azure AzureOperatorConfigSecretAzureOperatorSecretYamlServiceAzure
}

type AzureOperatorConfigSecretAzureOperatorSecretYamlServiceAzure struct {
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
	Template       AzureOperatorConfigSecretAzureOperatorSecretYamlServiceAzureTemplate
}

type AzureOperatorConfigSecretAzureOperatorSecretYamlServiceAzureTemplate struct {
	URI AzureOperatorConfigSecretAzureOperatorSecretYamlServiceAzureTemplateURI
}

type AzureOperatorConfigSecretAzureOperatorSecretYamlServiceAzureTemplateURI struct {
	// Version is currently the Github/CircleCI SHA.
	Version string
}

type AzureOperatorConfigSecretRegistry struct {
	PullSecret AzureOperatorConfigSecretRegistryPullSecret
}

type AzureOperatorConfigSecretRegistryPullSecret struct {
	DockerConfigJSON string
}

// NewAzureOperator renders values required by azure-operator-chart.
func NewAzureOperator(config AzureOperatorConfig) (string, error) {
	if config.Provider.Azure.Location == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Provider.Azure.Location must not be empty", config)
	}
	if config.Secret.AzureOperator.CredentialDefault.ClientID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.CredentialDefault.ClientID must not be empty", config)
	}
	if config.Secret.AzureOperator.CredentialDefault.ClientSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.CredentialDefault.ClientSecret must not be empty", config)
	}
	if config.Secret.AzureOperator.CredentialDefault.SubscriptionID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.CredentialDefault.SubscriptionID must not be empty", config)
	}
	if config.Secret.AzureOperator.CredentialDefault.TenantID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.CredentialDefault.TenantID must not be empty", config)
	}
	if config.Secret.AzureOperator.SecretYaml.Service.Azure.ClientID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.SecretYaml.Service.Azure.ClientID must not be empty", config)
	}
	if config.Secret.AzureOperator.SecretYaml.Service.Azure.ClientSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.SecretYaml.Service.Azure.ClientSecret must not be empty", config)
	}
	if config.Secret.AzureOperator.SecretYaml.Service.Azure.SubscriptionID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.SecretYaml.Service.Azure.SubscriptionID must not be empty", config)
	}
	if config.Secret.AzureOperator.SecretYaml.Service.Azure.TenantID == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.SecretYaml.Service.Azure.TenantID must not be empty", config)
	}
	if config.Secret.AzureOperator.SecretYaml.Service.Azure.Template.URI.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.AzureOperator.SecretYaml.Service.Azure.Template.URI.Version must not be empty", config)
	}
	if config.Secret.Registry.PullSecret.DockerConfigJSON == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Secret.Registry.PullSecret.DockerConfigJSON must not be empty", config)
	}

	values, err := render.Render(azureOperatorTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
