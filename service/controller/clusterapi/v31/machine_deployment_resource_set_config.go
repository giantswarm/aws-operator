package v30

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/cloudconfig"
	"github.com/giantswarm/aws-operator/service/locker"
)

type MachineDeploymentResourceSetConfig struct {
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
	OIDC                       cloudconfig.ConfigOIDC
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
