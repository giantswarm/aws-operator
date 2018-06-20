package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	azureclient "github.com/giantswarm/e2eclients/azure"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	VirtualMachineSize = "Standard_A2"
)

type AzureConfig struct {
	AzureClient    *azureclient.Client
	GuestFramework *framework.Guest
	HostFramework  *framework.Host
	Logger         micrologger.Logger

	ClusterID string
}

type Azure struct {
	azureClient    *azureclient.Client
	guestFramework *framework.Guest
	hostFramework  *framework.Host
	logger         micrologger.Logger

	clusterID string
}

func NewAzure(config AzureConfig) (*Azure, error) {
	if config.AzureClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AzureClient must not be empty", config)
	}
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
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
		azureClient:    config.AzureClient,
		guestFramework: config.GuestFramework,
		hostFramework:  config.HostFramework,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return a, nil
}

func (a *Azure) InstallTestApp() error {
	var err error

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: a.logger,

			Address:      CNRAddress,
			Organization: CNROrganization,
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:    a.logger,
			K8sClient: a.guestFramework.K8sClient(),

			RestConfig: a.guestFramework.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.EnsureTillerInstalled()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Install the e2e app chart in the guest cluster.
	{
		a.logger.Log("level", "debug", "message", "installing e2e-app for testing")

		tarballPath, err := apprClient.PullChartTarball(ChartName, ChartChannel)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.InstallFromTarball(tarballPath, ChartNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
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

func (a *Azure) WaitForAPIDown() error {
	err := a.guestFramework.WaitForAPIDown()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *Azure) WaitForGuestReady() error {
	err := a.guestFramework.WaitForGuestReady()
	if err != nil {
		return microerror.Mask(err)
	}

	// Wait for e2e app to be up.
	for {
		a.logger.Log("level", "debug", "message", "waiting for 2 pods of the e2e-app to be up")

		o := metav1.ListOptions{
			LabelSelector: "app=e2e-app",
		}
		l, err := a.guestFramework.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(l.Items) != 2 {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("found %d pods", len(l.Items)))
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	return nil
}
