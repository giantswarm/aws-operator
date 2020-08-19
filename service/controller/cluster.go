package controller

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	versionedinfrastructure "github.com/giantswarm/aws-operator/pkg/clientset/versioned"
	"github.com/giantswarm/aws-operator/service/controller/internal/adapter"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/network"
)

type ClusterConfig struct {
	CMAClient        clientset.Interface
	G8sClient        versioned.Interface
	G8sClientInfra   versionedinfrastructure.Interface
	K8sClient        kubernetes.Interface
	K8sExtClient     apiextensionsclient.Interface
	Logger           micrologger.Logger
	NetworkAllocator network.Allocator

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               FrameworkConfigAPIWhitelist
	DeleteLoggingBucket        bool
	EncrypterBackend           string
	GuestAWSConfig             ClusterConfigAWSConfig
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	GuestUpdateEnabled         bool
	HostAWSConfig              ClusterConfigAWSConfig
	IgnitionPath               string
	ImagePullProgressDeadline  string
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	LabelSelector              ClusterConfigLabelSelector
	OIDC                       ClusterConfigOIDC
	PodInfraContainerImage     string
	ProjectName                string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSOPublicKey               string
	VaultAddress               string
	VPCPeerID                  string
}

type ClusterConfigAWSConfig struct {
	AccessKeyID       string
	AccessKeySecret   string
	AvailabilityZones []string
	Region            string
	SessionToken      string
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

// FrameworkConfigAPIWhitelist defines guest cluster k8s API whitelisting types.
type FrameworkConfigAPIWhitelist struct {
	Private FrameworkConfigAPIWhitelistConfig
	Public  FrameworkConfigAPIWhitelistConfig
}

// FrameworkConfigAPIWhitelistConfig defines guest cluster k8s API whitelisting.
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
			Watcher: config.G8sClient.ProviderV1alpha1().AWSConfigs(""),

			ListOptions: metav1.ListOptions{
				LabelSelector: key.VersionLabelSelector(config.LabelSelector.Enabled, config.LabelSelector.OverridenVersion),
			},
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
			CRD:          v1alpha1.NewAWSConfigCRD(),
			CRDClient:    crdClient,
			Informer:     newInformer,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			RESTClient:   config.G8sClient.ProviderV1alpha1().RESTClient(),

			Name: config.ProjectName,
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

	var controlPlaneAWSClients awsclient.Clients
	{
		c := awsclient.Config{
			AccessKeyID:     config.HostAWSConfig.AccessKeyID,
			AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
			Region:          config.HostAWSConfig.Region,
			SessionToken:    config.HostAWSConfig.SessionToken,
		}

		controlPlaneAWSClients, err = awsclient.NewClients(c)
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

	var resourceSet *controller.ResourceSet
	{
		c := clusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			G8sClientInfra:         config.G8sClientInfra,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				Region:          config.HostAWSConfig.Region,
				SessionToken:    config.HostAWSConfig.SessionToken,
			},
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			NetworkAllocator:   config.NetworkAllocator,
			RandomKeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:  config.AccessLogsExpiration,
			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			APIWhitelist: adapter.APIWhitelist{
				Private: adapter.Whitelist{
					Enabled:    config.APIWhitelist.Private.Enabled,
					SubnetList: config.APIWhitelist.Private.SubnetList,
				},
				Public: adapter.Whitelist{
					Enabled:    config.APIWhitelist.Public.Enabled,
					SubnetList: config.APIWhitelist.Public.SubnetList,
				},
			},
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			ImagePullProgressDeadline:  config.ImagePullProgressDeadline,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
			VPCPeerID:      config.VPCPeerID,
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
