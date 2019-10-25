package cluster

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/cluster/adapter"
	"github.com/giantswarm/aws-operator/service/locker"
)

type ResourceSetConfig struct {
	CertsSearcher          certs.Interface
	CMAClient              clientset.Interface
	ControlPlaneAWSClients aws.Clients
	G8sClient              versioned.Interface
	HostAWSConfig          aws.Config
	K8sClient              kubernetes.Interface
	Locker                 locker.Interface
	Logger                 micrologger.Logger
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
	ImagePullProgressDeadline  string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	DeleteLoggingBucket        bool
	NetworkSetupDockerImage    string
	Route53Enabled             bool
	RouteTables                string
	PodInfraContainerImage     string
	RegistryDomain             string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
	VPCPeerID                  string
}

func (c ResourceSetConfig) GetEncrypterBackend() string {
	return c.EncrypterBackend
}

func (c ResourceSetConfig) GetInstallationName() string {
	return c.InstallationName
}

func (c ResourceSetConfig) GetLogger() micrologger.Logger {
	return c.Logger
}

func (c ResourceSetConfig) GetVaultAddress() string {
	return c.VaultAddress
}
