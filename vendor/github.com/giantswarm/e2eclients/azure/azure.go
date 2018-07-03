package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/giantswarm/microerror"
)

const (
	envVarAzureClientID       = "AZURE_CLIENTID"
	envVarAzureClientSecret   = "AZURE_CLIENTSECRET"
	envVarAzureSubscriptionID = "AZURE_SUBSCRIPTIONID"
	envVarAzureTenantID       = "AZURE_TENANTID"
)

var (
	azureClientID       string
	azureClientSecret   string
	azureSubscriptionID string
	azureTenantID       string
)

type Client struct {
	VirtualMachineScaleSetsClient *compute.VirtualMachineScaleSetsClient
}

func NewClient() (*Client, error) {
	a := &Client{}

	{
		azureClientID = os.Getenv(envVarAzureClientID)
		if azureClientID == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarAzureClientID)
		}

		azureClientSecret = os.Getenv(envVarAzureClientSecret)
		if azureClientSecret == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarAzureClientSecret)
		}

		azureSubscriptionID = os.Getenv(envVarAzureSubscriptionID)
		if azureSubscriptionID == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarAzureSubscriptionID)
		}

		azureTenantID = os.Getenv(envVarAzureTenantID)
		if azureTenantID == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarAzureTenantID)
		}

		env, err := azure.EnvironmentFromName(azure.PublicCloud.Name)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, azureTenantID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, azureClientID, azureClientSecret, env.ServiceManagementEndpoint)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		virtualMachineScaleSetsClient := compute.NewVirtualMachineScaleSetsClient(azureSubscriptionID)
		virtualMachineScaleSetsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)

		a.VirtualMachineScaleSetsClient = &virtualMachineScaleSetsClient
	}

	return a, nil
}
