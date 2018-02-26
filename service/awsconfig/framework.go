package awsconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeytpr"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/awsconfig/v1"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2"
	"github.com/giantswarm/aws-operator/service/awsconfig/v3"
	"github.com/giantswarm/aws-operator/service/awsconfig/v4"
	v4cloudconfig "github.com/giantswarm/aws-operator/service/awsconfig/v4/cloudconfig"
	"github.com/giantswarm/aws-operator/service/awsconfig/v5"
	v5cloudconfig "github.com/giantswarm/aws-operator/service/awsconfig/v5/cloudconfig"
	"github.com/giantswarm/aws-operator/service/awsconfig/v6"
	v6cloudconfig "github.com/giantswarm/aws-operator/service/awsconfig/v6/cloudconfig"
)

const (
	AWSConfigCleanupFinalizer = "aws-operator.giantswarm.io/custom-object-cleanup"
)

type FrameworkConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestAWSConfig     FrameworkConfigAWSConfig
	GuestUpdateEnabled bool
	HostAWSConfig      FrameworkConfigAWSConfig
	InstallationName   string
	OIDC               FrameworkConfigOIDCConfig
	// Name is the name of the project.
	Name       string
	PubKeyFile string
}

type FrameworkConfigAWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	SessionToken    string
}

// OIDC represents the configuration of the OIDC authorization provider
type FrameworkConfigOIDCConfig struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

func NewFramework(config FrameworkConfig) (*framework.Framework, error) {
	var err error

	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
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
	// config.GuestAWSConfig.SessionToken is optional.
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
		// config.HostAWSConfig.SessionToken is optional.
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = config.K8sExtClient
		c.Logger = config.Logger

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceRouter, err := newResourceRouter(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var newInformer *informer.Informer
	{
		c := informer.DefaultConfig()

		c.Watcher = config.G8sClient.ProviderV1alpha1().AWSConfigs("")

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var crdFramework *framework.Framework
	{
		c := framework.Config{
			CRD:            v1alpha1.NewAWSConfigCRD(),
			CRDClient:      crdClient,
			Informer:       newInformer,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,
		}

		crdFramework, err = framework.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}

func newResourceRouter(config FrameworkConfig) (*framework.ResourceRouter, error) {
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

	var resourceSetV1 *framework.ResourceSet
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
			ProjectName:      config.Name,
			PubKeyFile:       config.PubKeyFile,
		}

		resourceSetV1, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV2 *framework.ResourceSet
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
			ProjectName:      config.Name,
		}

		resourceSetV2, err = v2.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV3 *framework.ResourceSet
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
				// 2.0.2 fixes missing region in host account credentials, the change only affects service/framework.go
				"2.0.2",
			},
			InstallationName: config.InstallationName,
			ProjectName:      config.Name,
		}

		resourceSetV3, err = v3.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV4 *framework.ResourceSet
	{
		c := v4.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"2.1.0",
			},
			InstallationName: config.InstallationName,
			OIDC: v4cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.Name,
		}

		resourceSetV4, err = v4.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV5 *framework.ResourceSet
	{
		c := v5.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"2.1.1",
			},
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v5cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.Name,
		}

		resourceSetV5, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV6 *framework.ResourceSet
	{
		c := v6.ResourceSetConfig{
			CertsSearcher:      certWatcher,
			GuestAWSClients:    awsClients,
			HostAWSClients:     awsHostClients,
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			RandomkeysSearcher: keyWatcher,

			HandledVersionBundles: []string{
				"2.1.1",
			},
			GuestUpdateEnabled: config.GuestUpdateEnabled,
			InstallationName:   config.InstallationName,
			OIDC: v6cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName: config.Name,
		}

		resourceSetV6, err = v6.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *framework.ResourceRouter
	{
		c := framework.ResourceRouterConfig{
			ResourceSets: []*framework.ResourceSet{
				resourceSetV1,
				resourceSetV2,
				resourceSetV3,
				resourceSetV4,
				resourceSetV5,
				resourceSetV6,
			},
		}

		resourceRouter, err = framework.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceRouter, nil
}
