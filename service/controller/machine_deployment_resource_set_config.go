package controller

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/locker"
)

type machineDeploymentResourceSetConfig struct {
	CertsSearcher          certs.Interface
	CMAClient              clientset.Interface
	ControlPlaneAWSClients aws.Clients
	G8sClient              versioned.Interface
	K8sClient              kubernetes.Interface
	Locker                 locker.Interface
	Logger                 micrologger.Logger
	RandomKeysSearcher     randomkeys.Interface

	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DockerDaemonCIDR           string
	EncrypterBackend           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	HostAWSConfig              aws.Config
	IgnitionPath               string
	ImagePullProgressDeadline  string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	NetworkSetupDockerImage    string
	PodInfraContainerImage     string
	ProjectName                string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
	VPCPeerID                  string
}

func (c machineDeploymentResourceSetConfig) GetEncrypterBackend() string {
	return c.EncrypterBackend
}

func (c machineDeploymentResourceSetConfig) GetInstallationName() string {
	return c.InstallationName
}

func (c machineDeploymentResourceSetConfig) GetLogger() micrologger.Logger {
	return c.Logger
}

func (c machineDeploymentResourceSetConfig) GetVaultAddress() string {
	return c.VaultAddress
}
