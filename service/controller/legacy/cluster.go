package legacy

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
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	v22 "github.com/giantswarm/aws-operator/service/controller/legacy/v22"
	v22adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v22/adapter"
	v22cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v22/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v22patch1"
	v22patch1adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v22patch1/adapter"
	v22patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v22patch1/cloudconfig"
	v23 "github.com/giantswarm/aws-operator/service/controller/legacy/v23"
	v23adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v23/adapter"
	v23cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v23/cloudconfig"
	v24 "github.com/giantswarm/aws-operator/service/controller/legacy/v24"
	v24adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v24/adapter"
	v24cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v24/cloudconfig"
	v25 "github.com/giantswarm/aws-operator/service/controller/legacy/v25"
	v25adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v25/adapter"
	v25cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v25/cloudconfig"
	v26 "github.com/giantswarm/aws-operator/service/controller/legacy/v26"
	v26adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v26/adapter"
	v26cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v26/cloudconfig"
	v27 "github.com/giantswarm/aws-operator/service/controller/legacy/v27"
	v27adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v27/adapter"
	v27cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v27/cloudconfig"
	v28 "github.com/giantswarm/aws-operator/service/controller/legacy/v28"
	v28adapter "github.com/giantswarm/aws-operator/service/controller/legacy/v28/adapter"
	v28cloudconfig "github.com/giantswarm/aws-operator/service/controller/legacy/v28/cloudconfig"
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
	DeleteLoggingBucket        bool
	EncrypterBackend           string
	GuestAWSConfig             ClusterConfigAWSConfig
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	GuestUpdateEnabled         bool
	HostAWSConfig              ClusterConfigAWSConfig
	IgnitionPath               string
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
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
			Watcher: config.G8sClient.ProviderV1alpha1().AWSConfigs(""),

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

	var resourceSetV22 *controller.ResourceSet
	{
		c := v22.ClusterResourceSetConfig{
			CertsSearcher: certsSearcher,
			CMAClient:     config.CMAClient,
			G8sClient:     config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				SessionToken:    config.HostAWSConfig.SessionToken,
				Region:          config.HostAWSConfig.Region,
			},
			HostAWSClients:     controlPlaneAWSClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			NetworkAllocator:   config.NetworkAllocator,
			RandomKeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			GuestUpdateEnabled:         config.GuestUpdateEnabled,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v22cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v22adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.RouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV22, err = v22.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV22patch1 *controller.ResourceSet
	{
		c := v22patch1.ClusterResourceSetConfig{
			CertsSearcher: certsSearcher,
			CMAClient:     config.CMAClient,
			G8sClient:     config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				SessionToken:    config.HostAWSConfig.SessionToken,
				Region:          config.HostAWSConfig.Region,
			},
			HostAWSClients:     controlPlaneAWSClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			NetworkAllocator:   config.NetworkAllocator,
			RandomKeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			GuestUpdateEnabled:         config.GuestUpdateEnabled,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v22patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v22patch1adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.RouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV22patch1, err = v22patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV23 *controller.ResourceSet
	{
		c := v23.ClusterResourceSetConfig{
			CertsSearcher: certsSearcher,
			CMAClient:     config.CMAClient,
			G8sClient:     config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				SessionToken:    config.HostAWSConfig.SessionToken,
				Region:          config.HostAWSConfig.Region,
			},
			HostAWSClients:     controlPlaneAWSClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			NetworkAllocator:   config.NetworkAllocator,
			RandomKeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			GuestUpdateEnabled:         config.GuestUpdateEnabled,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v23cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v23adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.RouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV23, err = v23.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV24 *controller.ResourceSet
	{
		c := v24.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
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

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v24cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v24adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
		}

		resourceSetV24, err = v24.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV25 *controller.ResourceSet
	{
		c := v25.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
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

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v25cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v25adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
		}

		resourceSetV25, err = v25.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV26 *controller.ResourceSet
	{
		c := v26.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
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

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v26cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v26adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
		}

		resourceSetV26, err = v26.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV27 *controller.ResourceSet
	{
		c := v27.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
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

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v27cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v27adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
			VPCPeerID:      config.VPCPeerID,
		}

		resourceSetV27, err = v27.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV28 *controller.ResourceSet
	{
		c := v28.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
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

			AccessLogsExpiration:       config.AccessLogsExpiration,
			AdvancedMonitoringEC2:      config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v28cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v28adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
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
		resourceSetV22,
		resourceSetV22patch1,
		resourceSetV23,
		resourceSetV24,
		resourceSetV25,
		resourceSetV26,
		resourceSetV27,
		resourceSetV28,
	}

	return resourceSets, nil
}
