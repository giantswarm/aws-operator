package legacy

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
	"github.com/giantswarm/aws-operator/service/controller/legacy/v22"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v22patch1"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v23"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v24"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v25"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v26"
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

	var v22ResourceSet *controller.ResourceSet
	{
		c := v22.DrainerResourceSetConfig{
			G8sClient: config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				SessionToken:    config.HostAWSConfig.SessionToken,
				Region:          config.HostAWSConfig.Region,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}
		v22ResourceSet, err = v22.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v22patch1ResourceSet *controller.ResourceSet
	{
		c := v22patch1.DrainerResourceSetConfig{
			G8sClient: config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				SessionToken:    config.HostAWSConfig.SessionToken,
				Region:          config.HostAWSConfig.Region,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			GuestUpdateEnabled: config.GuestUpdateEnabled,
			ProjectName:        config.ProjectName,
		}
		v22patch1ResourceSet, err = v22patch1.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v23ResourceSet *controller.ResourceSet
	{
		c := v23.DrainerResourceSetConfig{
			G8sClient: config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				SessionToken:    config.HostAWSConfig.SessionToken,
				Region:          config.HostAWSConfig.Region,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

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
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				Region:          config.HostAWSConfig.Region,
				SessionToken:    config.HostAWSConfig.SessionToken,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ProjectName:    config.ProjectName,
			Route53Enabled: config.Route53Enabled,
		}

		v24ResourceSet, err = v24.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v25ResourceSet *controller.ResourceSet
	{
		c := v25.DrainerResourceSetConfig{
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				Region:          config.HostAWSConfig.Region,
				SessionToken:    config.HostAWSConfig.SessionToken,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ProjectName:    config.ProjectName,
			Route53Enabled: config.Route53Enabled,
		}

		v25ResourceSet, err = v25.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v26ResourceSet *controller.ResourceSet
	{
		c := v26.DrainerResourceSetConfig{
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				Region:          config.HostAWSConfig.Region,
				SessionToken:    config.HostAWSConfig.SessionToken,
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ProjectName:    config.ProjectName,
			Route53Enabled: config.Route53Enabled,
		}

		v26ResourceSet, err = v26.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		v22ResourceSet,
		v22patch1ResourceSet,
		v23ResourceSet,
		v24ResourceSet,
		v25ResourceSet,
		v26ResourceSet,
	}

	return resourceSets, nil
}
