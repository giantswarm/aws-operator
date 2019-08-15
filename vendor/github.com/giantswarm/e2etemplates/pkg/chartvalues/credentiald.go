package chartvalues

import (
	"reflect"

	"github.com/giantswarm/e2etemplates/internal/render"
	"github.com/giantswarm/microerror"
)

type CredentialdConfig struct {
	AWS                CredentialdConfigAWS
	Azure              CredentialdConfigAzure
	Deployment         CredentialdConfigDeployment
	RegistryPullSecret string
}

type CredentialdConfigAWS struct {
	CredentialDefault CredentialdConfigAWSCredentialDefault
}

type CredentialdConfigAWSCredentialDefault struct {
	AdminARN       string
	AWSOperatorARN string
}

type CredentialdConfigAzure struct {
	CredentialDefault CredentialdConfigAzureCredentialDefault
}

type CredentialdConfigDeployment struct {
	Replicas int
}

type CredentialdConfigAzureCredentialDefault struct {
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
}

func NewCredentiald(config CredentialdConfig) (string, error) {
	if config.RegistryPullSecret == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RegistryPullSecret must not be empty", config)
	}
	if reflect.DeepEqual(config.AWS, CredentialdConfigAWS{}) && reflect.DeepEqual(config.Azure, CredentialdConfigAzure{}) {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS or %T.Azure must not be empty", config, config)
	}
	if !reflect.DeepEqual(config.AWS, CredentialdConfigAWS{}) && !reflect.DeepEqual(config.Azure, CredentialdConfigAzure{}) {
		return "", microerror.Maskf(invalidConfigError, "%T.AWS and %T.Azure are mutually exclusive", config, config)
	}
	if !reflect.DeepEqual(config.AWS, CredentialdConfigAWS{}) {
		if config.AWS.CredentialDefault.AdminARN == "" {
			return "", microerror.Maskf(invalidConfigError, "%T.AWS.CredentialDefault.AdminARN must not be empty", config)
		}
		if config.AWS.CredentialDefault.AWSOperatorARN == "" {
			return "", microerror.Maskf(invalidConfigError, "%T.AWS.CredentialDefault.AWSOperatorARN must not be empty", config)
		}
	}
	if !reflect.DeepEqual(config.Azure, CredentialdConfigAzure{}) {
		if config.Azure.CredentialDefault.ClientID == "" {
			return "", microerror.Maskf(invalidConfigError, "%T.Azure.CredentialDefault.ClientID must not be empty", config)
		}
		if config.Azure.CredentialDefault.ClientSecret == "" {
			return "", microerror.Maskf(invalidConfigError, "%T.Azure.CredentialDefault.ClientSecret must not be empty", config)
		}
		if config.Azure.CredentialDefault.SubscriptionID == "" {
			return "", microerror.Maskf(invalidConfigError, "%T.Azure.CredentialDefault.SubscriptionID must not be empty", config)
		}
		if config.Azure.CredentialDefault.TenantID == "" {
			return "", microerror.Maskf(invalidConfigError, "%T.Azure.CredentialDefault.TenantID must not be empty", config)
		}
	}

	values, err := render.Render(credentialdTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
