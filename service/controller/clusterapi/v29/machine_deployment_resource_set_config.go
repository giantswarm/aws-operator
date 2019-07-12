package v29

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
)

type MachineDeploymentResourceSetConfig struct {
	CMAClient              clientset.Interface
	ControlPlaneAWSClients aws.Clients
	G8sClient              versioned.Interface
	K8sClient              kubernetes.Interface
	Logger                 micrologger.Logger

	EncrypterBackend           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	HostAWSConfig              aws.Config
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	ProjectName                string
	Route53Enabled             bool
	VaultAddress               string
}

func (c MachineDeploymentResourceSetConfig) GetEncrypterBackend() string {
	return c.EncrypterBackend
}

func (c MachineDeploymentResourceSetConfig) GetInstallationName() string {
	return c.InstallationName
}

func (c MachineDeploymentResourceSetConfig) GetLogger() micrologger.Logger {
	return c.Logger
}

func (c MachineDeploymentResourceSetConfig) GetVaultAddress() string {
	return c.VaultAddress
}
