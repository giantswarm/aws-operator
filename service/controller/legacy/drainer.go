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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/key"
	v25 "github.com/giantswarm/aws-operator/service/controller/legacy/v25"
	v26 "github.com/giantswarm/aws-operator/service/controller/legacy/v26"
	v27 "github.com/giantswarm/aws-operator/service/controller/legacy/v27"
	v28 "github.com/giantswarm/aws-operator/service/controller/legacy/v28"
	v28patch1 "github.com/giantswarm/aws-operator/service/controller/legacy/v28patch1"
	v29 "github.com/giantswarm/aws-operator/service/controller/legacy/v29"
	v30 "github.com/giantswarm/aws-operator/service/controller/legacy/v30"
)

type DrainerConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	GuestAWSConfig     DrainerConfigAWS
	GuestUpdateEnabled bool
	HostAWSConfig      DrainerConfigAWS
	LabelSelector      DrainerConfigLabelSelector
	ProjectName        string
	Route53Enabled     bool
}

type DrainerConfigAWS struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	SessionToken    string
}

type DrainerConfigLabelSelector struct {
	Enabled          bool
	OverridenVersion string
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

			ListOptions: metav1.ListOptions{
				LabelSelector: key.VersionLabelSelector(config.LabelSelector.Enabled, config.LabelSelector.OverridenVersion),
			},
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

	var v27ResourceSet *controller.ResourceSet
	{
		c := v27.DrainerResourceSetConfig{
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

		v27ResourceSet, err = v27.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v28ResourceSet *controller.ResourceSet
	{
		c := v28.DrainerResourceSetConfig{
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

		v28ResourceSet, err = v28.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v28patch1ResourceSet *controller.ResourceSet
	{
		c := v28patch1.DrainerResourceSetConfig{
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

		v28patch1ResourceSet, err = v28patch1.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v29ResourceSet *controller.ResourceSet
	{
		c := v29.DrainerResourceSetConfig{
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

		v29ResourceSet, err = v29.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v30ResourceSet *controller.ResourceSet
	{
		c := v30.DrainerResourceSetConfig{
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

		v30ResourceSet, err = v30.NewDrainerResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		v25ResourceSet,
		v26ResourceSet,
		v27ResourceSet,
		v28ResourceSet,
		v28patch1ResourceSet,
		v29ResourceSet,
		v30ResourceSet,
	}

	return resourceSets, nil
}
