package controller

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
	"github.com/giantswarm/aws-operator/service/controller/v22"
	"github.com/giantswarm/aws-operator/service/controller/v23"
	"github.com/giantswarm/aws-operator/service/controller/v24"
)

type DrainerConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestAWSConfig     DrainerConfigAWS
	GuestUpdateEnabled bool
	HostAWSConfig      DrainerConfigAWS
	ProjectName        string
	Route53Enabled     bool
}

type DrainerConfigAWS struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	SessionToken    string
}

type Drainer struct {
	*controller.Controller
}

func NewDrainer(config DrainerConfig) (*Drainer, error) {
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

	if config.GuestAWSConfig.AccessKeyID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestAWSConfig.AccessKeyID must not be empty", config)
	}
	if config.GuestAWSConfig.AccessKeySecret == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestAWSConfig.AccessKeySecret must not be empty", config)
	}
	if config.GuestAWSConfig.Region == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestAWSConfig.Region must not be empty", config)
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

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.G8sClient.ProviderV1alpha1().AWSConfigs(""),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: 30 * time.Second,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets, err := newDrainerResourceSets(config)
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

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/aws-operator-drainer.
			Name: config.ProjectName + "-drainer",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &Drainer{
		Controller: operatorkitController,
	}

	return d, nil
}

func newDrainerResourceSets(config DrainerConfig) ([]*controller.ResourceSet, error) {
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

	var v22ResourceSet *controller.ResourceSet
	{
		c := v22.DrainerResourceSetConfig{
			G8sClient:     config.G8sClient,
			HostAWSConfig: hostAWSConfig,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}
		v22ResourceSet, err = v22.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v23ResourceSet *controller.ResourceSet
	{
		c := v23.DrainerResourceSetConfig{
			G8sClient:     config.G8sClient,
			HostAWSConfig: hostAWSConfig,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}
		v23ResourceSet, err = v23.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v24ResourceSet *controller.ResourceSet
	{
		c := v24.DrainerResourceSetConfig{
			G8sClient:      config.G8sClient,
			HostAWSClients: awsHostClients,
			HostAWSConfig:  hostAWSConfig,
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
			Route53Enabled:     config.Route53Enabled,
		}

		v24ResourceSet, err = v24.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		v22ResourceSet,
		v23ResourceSet,
		v24ResourceSet,
	}

	return resourceSets, nil
}
