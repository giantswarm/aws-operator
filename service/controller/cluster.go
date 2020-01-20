package controller

import (
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/randomkeys"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp"
	"github.com/giantswarm/aws-operator/service/internal/locker"
)

type ClusterConfig struct {
	K8sClient k8sclient.Interface
	Locker    locker.Interface
	Logger    micrologger.Logger

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               ClusterConfigAPIWhitelist
	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DeleteLoggingBucket        bool
	DockerDaemonCIDR           string
	EncrypterBackend           string
	GuestAvailabilityZones     []string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	GuestUpdateEnabled         bool
	HostAWSConfig              aws.Config
	IgnitionPath               string
	ImagePullProgressDeadline  string
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	LabelSelector              ClusterConfigLabelSelector
	NetworkSetupDockerImage    string
	OIDC                       ClusterConfigOIDC
	PodInfraContainerImage     string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
}

type ClusterConfigLabelSelector struct {
	Enabled          bool
	OverridenVersion string
}

// ClusterConfigOIDC represents the configuration of the OIDC authorization
// provider.
type ClusterConfigOIDC struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

// ClusterConfigAPIWhitelist defines guest cluster k8s API whitelisting types.
type ClusterConfigAPIWhitelist struct {
	Private ClusterConfigAPIWhitelistConfig
	Public  ClusterConfigAPIWhitelistConfig
}

// ClusterConfigAPIWhitelistConfig defines guest cluster k8s API whitelisting.
type ClusterConfigAPIWhitelistConfig struct {
	Enabled    bool
	SubnetList string
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	resourceSets, err := newClusterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          infrastructurev1alpha2.NewAWSClusterCRD(),
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/aws-operator-cluster-controller.
			Name: project.Name() + "-cluster-controller",
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSCluster)
			},
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: operatorkitController,
	}

	return c, nil
}

func newClusterResourceSets(config ClusterConfig) ([]*controller.ResourceSet, error) {
	var err error

	var certsSearcher *certs.Searcher
	{
		c := certs.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var controlPlaneAWSClients aws.Clients
	{
		c := aws.Config{
			AccessKeyID:     config.HostAWSConfig.AccessKeyID,
			AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
			Region:          config.HostAWSConfig.Region,
			SessionToken:    config.HostAWSConfig.SessionToken,
		}

		controlPlaneAWSClients, err = aws.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomKeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomKeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSet *controller.ResourceSet
	{
		c := clusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			CtrlClient:             config.K8sClient.CtrlClient(),
			G8sClient:              config.K8sClient.G8sClient(),
			HostAWSConfig:          config.HostAWSConfig,
			K8sClient:              config.K8sClient.K8sClient(),
			Locker:                 config.Locker,
			Logger:                 config.Logger,
			RandomKeysSearcher:     randomKeysSearcher,

			AccessLogsExpiration:  config.AccessLogsExpiration,
			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			APIWhitelist: tccp.APIWhitelist{
				Private: tccp.Whitelist{
					Enabled:    config.APIWhitelist.Private.Enabled,
					SubnetList: config.APIWhitelist.Private.SubnetList,
				},
				Public: tccp.Whitelist{
					Enabled:    config.APIWhitelist.Public.Enabled,
					SubnetList: config.APIWhitelist.Public.SubnetList,
				},
			},
			CalicoCIDR:                 config.CalicoCIDR,
			CalicoMTU:                  config.CalicoMTU,
			CalicoSubnet:               config.CalicoSubnet,
			ClusterIPRange:             config.ClusterIPRange,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			DockerDaemonCIDR:           config.DockerDaemonCIDR,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			IgnitionPath:               config.IgnitionPath,
			ImagePullProgressDeadline:  config.ImagePullProgressDeadline,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			NetworkSetupDockerImage:    config.NetworkSetupDockerImage,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			RegistryDomain:             config.RegistryDomain,
			Route53Enabled:             config.Route53Enabled,
			RouteTables:                config.RouteTables,
			SSHUserList:                config.SSHUserList,
			SSOPublicKey:               config.SSOPublicKey,
			VaultAddress:               config.VaultAddress,
		}

		resourceSet, err = newClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
