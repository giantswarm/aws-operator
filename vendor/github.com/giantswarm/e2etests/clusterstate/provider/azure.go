package provider

import (
	"context"
	"fmt"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	azureclient "github.com/giantswarm/e2eclients/azure"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	VirtualMachineSize = "Standard_A2"
)

type AzureConfig struct {
	AzureClient   *azureclient.Client
	HostFramework *framework.Host
	Logger        micrologger.Logger

	ClusterID string
}

type Azure struct {
	azureClient   *azureclient.Client
	hostFramework *framework.Host
	logger        micrologger.Logger

	clusterID string
}

func NewAzure(config AzureConfig) (*Azure, error) {
	if config.AzureClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AzureClient must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	a := &Azure{
		azureClient:   config.AzureClient,
		hostFramework: config.HostFramework,
		logger:        config.Logger,

		clusterID: config.ClusterID,
	}

	return a, nil
}

func (a *Azure) RebootMaster() error {
	resourceGroupName := a.clusterID
	masterVMName := fmt.Sprintf("%s-Master-1", a.clusterID)
	_, err := a.azureClient.VirtualMachineClient.Restart(context.TODO(), resourceGroupName, masterVMName)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Azure) ReplaceMaster() error {
	customObject, err := a.hostFramework.G8sClient().ProviderV1alpha1().AzureConfigs("default").Get(a.clusterID, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	// Change virtual machine size to trigger replacement of existing master node.
	customObject.Spec.Azure.Masters[0].VMSize = VirtualMachineSize

	_, err = a.hostFramework.G8sClient().ProviderV1alpha1().AzureConfigs("default").Update(customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
