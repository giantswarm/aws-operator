package controller

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v17patch1"
	v17patch1adapter "github.com/giantswarm/aws-operator/service/controller/v17patch1/adapter"
	v17patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/v17patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v18"
	v18adapter "github.com/giantswarm/aws-operator/service/controller/v18/adapter"
	v18cloudconfig "github.com/giantswarm/aws-operator/service/controller/v18/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v19"
	v19adapter "github.com/giantswarm/aws-operator/service/controller/v19/adapter"
	v19cloudconfig "github.com/giantswarm/aws-operator/service/controller/v19/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v20"
	v20adapter "github.com/giantswarm/aws-operator/service/controller/v20/adapter"
	v20cloudconfig "github.com/giantswarm/aws-operator/service/controller/v20/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v21"
	v21adapter "github.com/giantswarm/aws-operator/service/controller/v21/adapter"
	v21cloudconfig "github.com/giantswarm/aws-operator/service/controller/v21/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v22"
	v22adapter "github.com/giantswarm/aws-operator/service/controller/v22/adapter"
	v22cloudconfig "github.com/giantswarm/aws-operator/service/controller/v22/cloudconfig"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/legacycerts/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

type ClusterConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

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
	PubKeyFile                 string
	PublicRouteTables          string
	RegistryDomain             string
	Route53Enabled             bool
	SSOPublicKey               string
	VaultAddress               string
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
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.K8sExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sExtClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	if config.GuestAWSConfig.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.AccessKeyID must not be empty")
	}
	if config.GuestAWSConfig.AccessKeySecret == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.AccessKeySecret must not be empty")
	}
	if config.GuestAWSConfig.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.GuestAWSConfig.Region must not be empty")
	}
	// TODO: remove this when all version prior to v11 are removed
	if config.HostAWSConfig.AccessKeyID == "" && config.HostAWSConfig.AccessKeySecret == "" {
		config.Logger.Log("debug", "no host cluster account credentials supplied, assuming guest and host uses same account")
		config.HostAWSConfig = config.GuestAWSConfig
	} else {
		if config.HostAWSConfig.AccessKeyID == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.AccessKeyID must not be empty")
		}
		if config.HostAWSConfig.AccessKeySecret == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.AccessKeySecret must not be empty")
		}
		if config.HostAWSConfig.Region == "" {
			return nil, microerror.Maskf(invalidConfigError, "config.HostAWSConfig.Region must not be empty")
		}
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
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

	hostAWSConfig := awsclient.Config{
		AccessKeyID:     config.HostAWSConfig.AccessKeyID,
		AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
		SessionToken:    config.HostAWSConfig.SessionToken,
		Region:          config.HostAWSConfig.Region,
	}

	awsHostClients, err := awsclient.NewClients(hostAWSConfig)
	if err != nil {
		return nil, microerror.Mask(err)
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

	var legacyCertsSearcher *legacy.Service
	{
		certConfig := legacy.DefaultServiceConfig()
		certConfig.K8sClient = config.K8sClient
		certConfig.Logger = config.Logger
		legacyCertsSearcher, err = legacy.NewService(certConfig)
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

	var resourceSetV17Patch1 *controller.ResourceSet
	{
		c := v17patch1.ClusterResourceSetConfig{
			CertsSearcher:      legacyCertsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:   config.AccessLogsExpiration,
			AdvancedMonitoringEC2:  config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:    config.DeleteLoggingBucket,
			EncrypterBackend:       config.EncrypterBackend,
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v17patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v17patch1adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV17Patch1, err = v17patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV18 *controller.ResourceSet
	{
		c := v18.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomKeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:   config.AccessLogsExpiration,
			AdvancedMonitoringEC2:  config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:    config.DeleteLoggingBucket,
			EncrypterBackend:       config.EncrypterBackend,
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v18cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v18adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV18, err = v18.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV19 *controller.ResourceSet
	{
		c := v19.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
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
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v19cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v19adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV19, err = v19.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV20 *controller.ResourceSet
	{
		c := v20.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
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
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: v20cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v20adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV20, err = v20.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV21 *controller.ResourceSet
	{
		c := v21.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
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
			OIDC: v21cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v21adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV21, err = v21.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV22 *controller.ResourceSet
	{
		c := v22.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
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
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV22, err = v22.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSetV17Patch1,
		resourceSetV18,
		resourceSetV19,
		resourceSetV20,
		resourceSetV21,
		resourceSetV22,
	}

	return resourceSets, nil
}
