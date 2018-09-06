package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	"github.com/giantswarm/randomkeytpr"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v1"
	"github.com/giantswarm/aws-operator/service/controller/v12"
	v12adapter "github.com/giantswarm/aws-operator/service/controller/v12/adapter"
	v12cloudconfig "github.com/giantswarm/aws-operator/service/controller/v12/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1"
	v12patch1adapter "github.com/giantswarm/aws-operator/service/controller/v12patch1/adapter"
	v12patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/v12patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v13"
	v13adapter "github.com/giantswarm/aws-operator/service/controller/v13/adapter"
	v13cloudconfig "github.com/giantswarm/aws-operator/service/controller/v13/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v14"
	v14adapter "github.com/giantswarm/aws-operator/service/controller/v14/adapter"
	v14cloudconfig "github.com/giantswarm/aws-operator/service/controller/v14/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v14patch1"
	v14patch1adapter "github.com/giantswarm/aws-operator/service/controller/v14patch1/adapter"
	v14patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/v14patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v15"
	v15adapter "github.com/giantswarm/aws-operator/service/controller/v15/adapter"
	v15cloudconfig "github.com/giantswarm/aws-operator/service/controller/v15/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v16"
	v16adapter "github.com/giantswarm/aws-operator/service/controller/v16/adapter"
	v16cloudconfig "github.com/giantswarm/aws-operator/service/controller/v16/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v17"
	v17adapter "github.com/giantswarm/aws-operator/service/controller/v17/adapter"
	v17cloudconfig "github.com/giantswarm/aws-operator/service/controller/v17/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v2"
	"github.com/giantswarm/aws-operator/service/controller/v3"
	"github.com/giantswarm/aws-operator/service/controller/v6"
	v6cloudconfig "github.com/giantswarm/aws-operator/service/controller/v6/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v8"
	v8cloudconfig "github.com/giantswarm/aws-operator/service/controller/v8/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v9patch1"
	v9patch1adapter "github.com/giantswarm/aws-operator/service/controller/v9patch1/adapter"
	v9patch1cloudconfig "github.com/giantswarm/aws-operator/service/controller/v9patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v9patch2"
	v9patch2adapter "github.com/giantswarm/aws-operator/service/controller/v9patch2/adapter"
	v9patch2cloudconfig "github.com/giantswarm/aws-operator/service/controller/v9patch2/cloudconfig"
)

type ClusterConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	AccessLogsExpiration   int
	AdvancedMonitoringEC2  bool
	APIWhitelist           FrameworkConfigAPIWhitelistConfig
	DeleteLoggingBucket    bool
	EncrypterBackend       string
	GuestAWSConfig         ClusterConfigAWSConfig
	GuestUpdateEnabled     bool
	HostAWSConfig          ClusterConfigAWSConfig
	IncludeTags            bool
	InstallationName       string
	OIDC                   ClusterConfigOIDC
	PodInfraContainerImage string
	ProjectName            string
	PubKeyFile             string
	PublicRouteTables      string
	RegistryDomain         string
	Route53Enabled         bool
	SSOPublicKey           string
	VaultAddress           string
}

type ClusterConfigAWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	SessionToken    string
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

	guestAWSConfig := awsclient.Config{
		AccessKeyID:     config.GuestAWSConfig.AccessKeyID,
		AccessKeySecret: config.GuestAWSConfig.AccessKeySecret,
		SessionToken:    config.GuestAWSConfig.SessionToken,
		Region:          config.GuestAWSConfig.Region,
	}

	hostAWSConfig := awsclient.Config{
		AccessKeyID:     config.HostAWSConfig.AccessKeyID,
		AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
		SessionToken:    config.HostAWSConfig.SessionToken,
		Region:          config.HostAWSConfig.Region,
	}

	awsClients := awsclient.NewClients(guestAWSConfig)

	awsHostClients := awsclient.NewClients(hostAWSConfig)

	var certsSearcher *legacy.Service
	{
		certConfig := legacy.DefaultServiceConfig()
		certConfig.K8sClient = config.K8sClient
		certConfig.Logger = config.Logger
		certsSearcher, err = legacy.NewService(certConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keyWatcher *randomkeytpr.Service
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = config.K8sClient
		keyConfig.Logger = config.Logger
		keyWatcher, err = randomkeytpr.NewService(keyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var randomKeySearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		randomKeySearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV1 *controller.ResourceSet
	{
		c := v1.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSConfig:     guestAWSConfig,
			HostAWSConfig:      hostAWSConfig,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"",
				"0.1.0",
				"1.0.0",
			},
			InstallationName: config.InstallationName,
			ProjectName:      config.ProjectName,
			PubKeyFile:       config.PubKeyFile,
		}

		resourceSetV1, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV2 *controller.ResourceSet
	{
		c := v2.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"0.2.0",
				"2.0.0",
			},
			InstallationName: config.InstallationName,
			ProjectName:      config.ProjectName,
		}

		resourceSetV2, err = v2.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV3 *controller.ResourceSet
	{
		c := v3.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"2.0.1",
				// 2.0.2 fixes missing region in host account credentials, the change only affects service/controller.go
				"2.0.2",
			},
			InstallationName: config.InstallationName,
			ProjectName:      config.ProjectName,
		}

		resourceSetV3, err = v3.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV6 *controller.ResourceSet
	{
		c := v6.ResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v6cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV6, err = v6.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV8 *controller.ResourceSet
	{
		c := v8.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v8cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV8, err = v8.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV9Patch1 *controller.ResourceSet
	{
		c := v9patch1.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration: config.AccessLogsExpiration,
			GuestUpdateEnabled:   config.GuestUpdateEnabled,
			InstallationName:     config.InstallationName,
			OIDC: v9patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v9patch1adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV9Patch1, err = v9patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV9Patch2 *controller.ResourceSet
	{
		c := v9patch2.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration: config.AccessLogsExpiration,
			GuestUpdateEnabled:   config.GuestUpdateEnabled,
			InstallationName:     config.InstallationName,
			OIDC: v9patch2cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v9patch2adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV9Patch2, err = v9patch2.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV12 *controller.ResourceSet
	{
		c := v12.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

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
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

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
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

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

	var resourceSetV14 *controller.ResourceSet
	{
		c := v14.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration:   config.AccessLogsExpiration,
			AdvancedMonitoringEC2:  config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:    config.DeleteLoggingBucket,
			EncrypterBackend:       config.EncrypterBackend,
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v14cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v14adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV14, err = v14.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV14Patch1 *controller.ResourceSet
	{
		c := v14patch1.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration:   config.AccessLogsExpiration,
			AdvancedMonitoringEC2:  config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:    config.DeleteLoggingBucket,
			EncrypterBackend:       config.EncrypterBackend,
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v14patch1cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v14patch1adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV14Patch1, err = v14patch1.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV15 *controller.ResourceSet
	{
		c := v15.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration:   config.AccessLogsExpiration,
			AdvancedMonitoringEC2:  config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:    config.DeleteLoggingBucket,
			EncrypterBackend:       config.EncrypterBackend,
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v15cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v15adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV15, err = v15.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV16 *controller.ResourceSet
	{
		c := v16.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration:   config.AccessLogsExpiration,
			AdvancedMonitoringEC2:  config.AdvancedMonitoringEC2,
			DeleteLoggingBucket:    config.DeleteLoggingBucket,
			EncrypterBackend:       config.EncrypterBackend,
			GuestUpdateEnabled:     config.GuestUpdateEnabled,
			PodInfraContainerImage: config.PodInfraContainerImage,
			Route53Enabled:         config.Route53Enabled,
			IncludeTags:            config.IncludeTags,
			InstallationName:       config.InstallationName,
			OIDC: v16cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v16adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName:       config.ProjectName,
			PublicRouteTables: config.PublicRouteTables,
			RegistryDomain:    config.RegistryDomain,
			SSOPublicKey:      config.SSOPublicKey,
			VaultAddress:      config.VaultAddress,
		}

		resourceSetV16, err = v16.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV17 *controller.ResourceSet
	{
		c := v17.ClusterResourceSetConfig{
			CertsSearcher:      certsSearcher,
			G8sClient:          config.G8sClient,
			HostAWSConfig:      hostAWSConfig,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

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

	resourceSets := []*controller.ResourceSet{
		resourceSetV1,
		resourceSetV2,
		resourceSetV3,
		resourceSetV6,
		resourceSetV8,
		resourceSetV9Patch1,
		resourceSetV9Patch2,
		resourceSetV12,
		resourceSetV12Patch1,
		resourceSetV13,
		resourceSetV14,
		resourceSetV14Patch1,
		resourceSetV15,
		resourceSetV16,
		resourceSetV17,
	}

	return resourceSets, nil
}
