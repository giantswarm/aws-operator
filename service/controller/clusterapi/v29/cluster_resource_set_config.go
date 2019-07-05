package v29

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/adapter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/cloudconfig"
	"github.com/giantswarm/aws-operator/service/network"
)

type ClusterResourceSetConfig struct {
	CertsSearcher          certs.Interface
	CMAClient              clientset.Interface
	ControlPlaneAWSClients aws.Clients
	G8sClient              versioned.Interface
	HostAWSConfig          aws.Config
	K8sClient              kubernetes.Interface
	Logger                 micrologger.Logger
	NetworkAllocator       network.Allocator
	RandomKeysSearcher     randomkeys.Interface

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               adapter.APIWhitelist
	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DockerDaemonCIDR           string
	EncrypterBackend           string
	GuestAvailabilityZones     []string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	IncludeTags                bool
	IgnitionPath               string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	DeleteLoggingBucket        bool
	NetworkSetupDockerImage    string
	OIDC                       cloudconfig.ConfigOIDC
	ProjectName                string
	Route53Enabled             bool
	RouteTables                string
	PodInfraContainerImage     string
	RegistryDomain             string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
	VPCPeerID                  string
}

func (c ClusterResourceSetConfig) GetEncrypterBackend() string {
	return c.EncrypterBackend
}

func (c ClusterResourceSetConfig) GetInstallationName() string {
	return c.InstallationName
}

func (c ClusterResourceSetConfig) GetLogger() micrologger.Logger {
	return c.Logger
}

func (c ClusterResourceSetConfig) GetVaultAddress() string {
	return c.VaultAddress
}
