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
	"github.com/giantswarm/aws-operator/service/controller/v10"
	v10adapter "github.com/giantswarm/aws-operator/service/controller/v10/adapter"
	v10cloudconfig "github.com/giantswarm/aws-operator/service/controller/v10/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v11"
	v11adapter "github.com/giantswarm/aws-operator/service/controller/v11/adapter"
	v11cloudconfig "github.com/giantswarm/aws-operator/service/controller/v11/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v12"
	v12adapter "github.com/giantswarm/aws-operator/service/controller/v12/adapter"
	v12cloudconfig "github.com/giantswarm/aws-operator/service/controller/v12/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v13"
	v13adapter "github.com/giantswarm/aws-operator/service/controller/v13/adapter"
	v13cloudconfig "github.com/giantswarm/aws-operator/service/controller/v13/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v2"
	"github.com/giantswarm/aws-operator/service/controller/v3"
	"github.com/giantswarm/aws-operator/service/controller/v4"
	v4cloudconfig "github.com/giantswarm/aws-operator/service/controller/v4/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v5"
	v5cloudconfig "github.com/giantswarm/aws-operator/service/controller/v5/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v6"
	v6cloudconfig "github.com/giantswarm/aws-operator/service/controller/v6/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v7"
	v7cloudconfig "github.com/giantswarm/aws-operator/service/controller/v7/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v8"
	v8cloudconfig "github.com/giantswarm/aws-operator/service/controller/v8/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v9"
	v9cloudconfig "github.com/giantswarm/aws-operator/service/controller/v9/cloudconfig"
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

	resourceRouter, err := newClusterResourceRouter(config)
	if err != nil {
		return nil, microerror.Mask(err)
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

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:            v1alpha1.NewAWSConfigCRD(),
			CRDClient:      crdClient,
			Informer:       newInformer,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,
			RESTClient:     config.G8sClient.ProviderV1alpha1().RESTClient(),

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

func newClusterResourceRouter(config ClusterConfig) (*controller.ResourceRouter, error) {
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

	var certWatcher *legacy.Service
	{
		certConfig := legacy.DefaultServiceConfig()
		certConfig.K8sClient = config.K8sClient
		certConfig.Logger = config.Logger
		certWatcher, err = legacy.NewService(certConfig)
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
			CertsSearcher:      certWatcher,
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
			CertsSearcher:      certWatcher,
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
			CertsSearcher:      certWatcher,
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

	var resourceSetV4 *controller.ResourceSet
	{
		c := v4.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			InstallationName: config.InstallationName,
			OIDC: v4cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV4, err = v4.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV5 *controller.ResourceSet
	{
		c := v5.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v5cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV5, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV6 *controller.ResourceSet
	{
		c := v6.ResourceSetConfig{
			CertsSearcher:      certWatcher,
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

	var resourceSetV7 *controller.ResourceSet
	{
		c := v7.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v7cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV7, err = v7.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV8 *controller.ResourceSet
	{
		c := v8.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
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

	var resourceSetV9 *controller.ResourceSet
	{
		c := v9.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v9cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV9, err = v9.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV9Patch1 *controller.ResourceSet
	{
		c := v9patch1.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
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
			CertsSearcher:      certWatcher,
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

	var resourceSetV10 *controller.ResourceSet
	{
		c := v10.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: randomKeySearcher,

			AccessLogsExpiration: config.AccessLogsExpiration,
			GuestUpdateEnabled:   config.GuestUpdateEnabled,
			InstallationName:     config.InstallationName,
			OIDC: v10cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v10adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV10, err = v10.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV11 *controller.ResourceSet
	{
		c := v11.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
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
			OIDC: v11cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			APIWhitelist: v11adapter.APIWhitelist{
				Enabled:    config.APIWhitelist.Enabled,
				SubnetList: config.APIWhitelist.SubnetList,
			},
			ProjectName: config.ProjectName,
		}

		resourceSetV11, err = v11.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV12 *controller.ResourceSet
	{
		c := v12.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
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

	var resourceSetV13 *controller.ResourceSet
	{
		c := v13.ClusterResourceSetConfig{
			CertsSearcher:      certWatcher,
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
			ProjectName:  config.ProjectName,
			SSOPublicKey: config.SSOPublicKey,
			VaultAddress: config.VaultAddress,
		}

		resourceSetV13, err = v13.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *controller.ResourceRouter
	{
		c := controller.ResourceRouterConfig{
			Logger: config.Logger,

			ResourceSets: []*controller.ResourceSet{
				resourceSetV1,
				resourceSetV2,
				resourceSetV3,
				resourceSetV4,
				resourceSetV5,
				resourceSetV6,
				resourceSetV7,
				resourceSetV8,
				resourceSetV9,
				resourceSetV9Patch1,
				resourceSetV9Patch2,
				resourceSetV10,
				resourceSetV11,
				resourceSetV12,
				resourceSetV13,
			},
		}

		resourceRouter, err = controller.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceRouter, nil
}
