package controller

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v12"
	v12adapter "github.com/giantswarm/aws-operator/service/controller/v12/adapter"
	v12cloudconfig "github.com/giantswarm/aws-operator/service/controller/v12/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1"
	v12patch1adapter "github.com/giantswarm/aws-operator/service/controller/v12patch1/adapter"
	v12patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/v12patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v13"
	v13adapter "github.com/giantswarm/aws-operator/service/controller/v13/adapter"
	v13cloudconfig "github.com/giantswarm/aws-operator/service/controller/v13/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3"
	v14patch3adapter "github.com/giantswarm/aws-operator/service/controller/v14patch3/adapter"
	v14patch3cloudconfig "github.com/giantswarm/aws-operator/service/controller/v14patch3/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v14patch4"
	v14patch4adapter "github.com/giantswarm/aws-operator/service/controller/v14patch4/adapter"
	v14patch4cloudconfig "github.com/giantswarm/aws-operator/service/controller/v14patch4/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1"
	v16patch1adapter "github.com/giantswarm/aws-operator/service/controller/v16patch1/adapter"
	v16patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/v16patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v17"
	v17adapter "github.com/giantswarm/aws-operator/service/controller/v17/adapter"
	v17cloudconfig "github.com/giantswarm/aws-operator/service/controller/v17/cloudconfig"
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

	var resourceSetV12 *controller.ResourceSet
	{
		c := v12.ClusterResourceSetConfig{
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
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v12cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v12adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV12, err = v12.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV12Patch1 *controller.ResourceSet
	{
		c := v12patch1.ClusterResourceSetConfig{
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
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v12patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v12patch1adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV12Patch1, err = v12patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV13 *controller.ResourceSet
	{
		c := v13.ClusterResourceSetConfig{
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
			OIDC: v13cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v13adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV13, err = v13.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch3 *controller.ResourceSet
	{
		c := v14patch3.ClusterResourceSetConfig{
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
			OIDC: v14patch3cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v14patch3adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV14Patch3, err = v14patch3.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch4 *controller.ResourceSet
	{
		c := v14patch4.ClusterResourceSetConfig{
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
			OIDC: v14patch4cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v14patch4adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV14Patch4, err = v14patch4.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV16Patch1 *controller.ResourceSet
	{
		c := v16patch1.ClusterResourceSetConfig{
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
			OIDC: v16patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v16patch1adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV16Patch1, err = v16patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV17 *controller.ResourceSet
	{
		c := v17.ClusterResourceSetConfig{
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
			OIDC: v17cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v17adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV17, err = v17.NewClusterResourceSet(c)
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

	resourceSets := []*controller.ResourceSet{
		resourceSetV12,
		resourceSetV12Patch1,
		resourceSetV13,
		resourceSetV14Patch3,
		resourceSetV14Patch4,
		resourceSetV16Patch1,
		resourceSetV17,
		resourceSetV17Patch1,
		resourceSetV18,
		resourceSetV19,
		resourceSetV20,
	}

	return resourceSets, nil
}
