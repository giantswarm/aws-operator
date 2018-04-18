package awsconfig

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v10"
	"github.com/giantswarm/aws-operator/service/controller/v7"
	"github.com/giantswarm/aws-operator/service/controller/v8"
	"github.com/giantswarm/aws-operator/service/controller/v9"
)

type DrainerFrameworkConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	AWS                DrainerFrameworkConfigAWS
	GuestUpdateEnabled bool
	ProjectName        string
}

type DrainerFrameworkConfigAWS struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	SessionToken    string
}

func NewDrainerFramework(config DrainerFrameworkConfig) (*controller.Controller, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.K8sExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sExtClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.AWS.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWS.AccessKeyID must not be empty", config)
	}
	if config.AWS.AccessKeySecret == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWS.AccessKeySecret must not be empty", config)
	}
	if config.AWS.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWS.Region must not be empty", config)
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var err error

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

	resourceRouter, err := newDrainerResourceRouter(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Watcher: config.G8sClient.ProviderV1alpha1().AWSConfigs(""),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: 30 * time.Second,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var crdFramework *controller.Controller
	{
		c := controller.Config{
			CRD:            v1alpha1.NewAWSConfigCRD(),
			CRDClient:      crdClient,
			Informer:       newInformer,
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/aws-operator-drainer.
			Name: config.ProjectName + "-drainer",
		}

		crdFramework, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return crdFramework, nil
}

func newDrainerResourceRouter(config DrainerFrameworkConfig) (*controller.ResourceRouter, error) {
	var err error

	var awsClients awsclient.Clients
	{
		c := awsclient.Config{
			AccessKeyID:     config.AWS.AccessKeyID,
			AccessKeySecret: config.AWS.AccessKeySecret,
			SessionToken:    config.AWS.SessionToken,
			Region:          config.AWS.Region,
		}

		awsClients = awsclient.NewClients(c)
	}

	var v7ResourceSet *controller.ResourceSet
	{
		c := v7.DrainerResourceSetConfig{
			GuestAWSClients: awsClients,
			Logger:          config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		v7ResourceSet, err = v7.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v8ResourceSet *controller.ResourceSet
	{
		c := v8.DrainerResourceSetConfig{
			AWS:       awsClients,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		v8ResourceSet, err = v8.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v9ResourceSet *controller.ResourceSet
	{
		c := v9.DrainerResourceSetConfig{
			AWS:       awsClients,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		v9ResourceSet, err = v9.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v10ResourceSet *controller.ResourceSet
	{
		c := v10.DrainerResourceSetConfig{
			AWS:       awsClients,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}

		v10ResourceSet, err = v10.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *controller.ResourceRouter
	{
		c := controller.ResourceRouterConfig{
			Logger: config.Logger,

			ResourceSets: []*controller.ResourceSet{
				v7ResourceSet,
				v8ResourceSet,
				v9ResourceSet,
				v10ResourceSet,
			},
		}

		resourceRouter, err = controller.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceRouter, nil
}
