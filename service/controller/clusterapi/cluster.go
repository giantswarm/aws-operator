package clusterapi

import (
	"net"

	clusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28"
	v28adapter "github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/adapter"
	v28cloudconfig "github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/cloudconfig"
	"github.com/giantswarm/aws-operator/service/network"
)

type ClusterConfig struct {
	CMAClient        clientset.Interface
	G8sClient        versioned.Interface
	K8sClient        kubernetes.Interface
	K8sExtClient     apiextensionsclient.Interface
	Logger           micrologger.Logger
	NetworkAllocator network.Allocator

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               FrameworkConfigAPIWhitelistConfig
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
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	NetworkSetupDockerImage    string
	OIDC                       ClusterConfigOIDC
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

// ClusterConfigOIDC represents the configuration of the OIDC authorization
// provider.
type ClusterConfigOIDC struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

// Whitelist defines guest cluster k8s API whitelisting.
type FrameworkConfigAPIWhitelistConfig struct {
	Enabled    bool
	SubnetList string
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: config.K8sExtClient,
			Logger:       config.Logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.CMAClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets, err := newClusterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          clusterv1alpha1.NewClusterCRD(),
			CRDClient:    crdClient,
			Informer:     newInformer,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			RESTClient:   config.CMAClient.ClusterV1alpha1().RESTClient(),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/aws-operator-cluster-controller.
			Name: config.ProjectName + "-cluster-controller",
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

	var certsSearcher *certs.Searcher
	{
		c := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomKeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		randomKeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV28 *controller.ResourceSet
	{
		c := v28.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			CMAClient:              config.CMAClient,
			G8sClient:              config.G8sClient,
			HostAWSConfig:          config.HostAWSConfig,
			K8sClient:              config.K8sClient,
			Logger:                 config.Logger,
			NetworkAllocator:       config.NetworkAllocator,
			RandomKeysSearcher:     randomKeysSearcher,

			AccessLogsExpiration:  config.AccessLogsExpiration,
			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			APIWhitelist: v28adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
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
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			NetworkSetupDockerImage:    config.NetworkSetupDockerImage,
			OIDC: v28cloudconfig.ConfigOIDC{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSHUserList:    config.SSHUserList,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
			VPCPeerID:      config.VPCPeerID,
		}

		resourceSetV28, err = v28.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSetV28,
	}

	return resourceSets, nil
}
