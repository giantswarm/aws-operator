package controller

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/service/internal/locker"
)

type clusterResourceSetConfig struct {
	CertsSearcher      certs.Interface
	CtrlClient         client.Client
	G8sClient          versioned.Interface
	HostAWSConfig      aws.Config
	K8sClient          kubernetes.Interface
	Locker             locker.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               tccp.APIWhitelist
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
	KubernetesDomain           string
	NetworkSetupDockerImage    string
	Route53Enabled             bool
	RouteTables                string
	PodInfraContainerImage     string
	RegistryDomain             string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
}

func (c clusterResourceSetConfig) GetEncrypterBackend() string {
	return c.EncrypterBackend
}

func (c clusterResourceSetConfig) GetInstallationName() string {
	return c.InstallationName
}

func (c clusterResourceSetConfig) GetLogger() micrologger.Logger {
	return c.Logger
}

func (c clusterResourceSetConfig) GetVaultAddress() string {
	return c.VaultAddress
}
