package controller

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/randomkeys"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
)

type ControlPlaneConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	APIWhitelist              ClusterConfigAPIWhitelist
	CalicoCIDR                int
	CalicoMTU                 int
	CalicoSubnet              string
	ClusterDomain             string
	ClusterIPRange            string
	DockerDaemonCIDR          string
	ExternalSNAT              bool
	HostAWSConfig             aws.Config
	IgnitionPath              string
	ImagePullProgressDeadline string
	InstallationName          string
	NetworkSetupDockerImage   string
	PodInfraContainerImage    string
	RegistryDomain            string
	Route53Enabled            bool
	SSHUserList               string
	SSOPublicKey              string
	VaultAddress              string
}

type ControlPlane struct {
	*controller.Controller
}

func NewControlPlane(config ControlPlaneConfig) (*ControlPlane, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resourceSets, err := newControlPlaneResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-control-plane-controller.
			Name: project.Name() + "-control-plane-controller",
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSControlPlane)
			},
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	d := &ControlPlane{
		Controller: operatorkitController,
	}

	return d, nil
}

func newControlPlaneResourceSets(config ControlPlaneConfig) ([]*controller.ResourceSet, error) {
	var err error

	var certsSearcher *certs.Searcher
	{
		c := certs.Config{
			K8sClient: config.K8sClient.K8sClient(),
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
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomKeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSet *controller.ResourceSet
	{
		c := controlPlaneResourceSetConfig{
			G8sClient:          config.K8sClient.G8sClient(),
			K8sClient:          config.K8sClient.K8sClient(),
			Logger:             config.Logger,
			CertsSearcher:      certsSearcher,
			RandomKeysSearcher: randomKeysSearcher,

			CalicoCIDR:                config.CalicoCIDR,
			CalicoMTU:                 config.CalicoMTU,
			CalicoSubnet:              config.CalicoSubnet,
			ClusterDomain:             config.ClusterDomain,
			ClusterIPRange:            config.ClusterIPRange,
			DockerDaemonCIDR:          config.DockerDaemonCIDR,
			IgnitionPath:              config.IgnitionPath,
			ImagePullProgressDeadline: config.ImagePullProgressDeadline,
			InstallationName:          config.InstallationName,
			HostAWSConfig:             config.HostAWSConfig,
			NetworkSetupDockerImage:   config.NetworkSetupDockerImage,
			PodInfraContainerImage:    config.PodInfraContainerImage,
			RegistryDomain:            config.RegistryDomain,
			Route53Enabled:            config.Route53Enabled,
			SSHUserList:               config.SSHUserList,
			SSOPublicKey:              config.SSOPublicKey,
			VaultAddress:              config.VaultAddress,
		}

		resourceSet, err = newControlPlaneResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
