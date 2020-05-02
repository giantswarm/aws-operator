package controller

import (
	"net"

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
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/images"
	"github.com/giantswarm/aws-operator/service/internal/locker"
)

type MachineDeploymentConfig struct {
	CertsSearcher      certs.Interface
	HAMaster           hamaster.Interface
	Images             images.Interface
	K8sClient          k8sclient.Interface
	Locker             locker.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DockerDaemonCIDR           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	HostAWSConfig              aws.Config
	IgnitionPath               string
	ImagePullProgressDeadline  string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	ClusterDomain              string
	NetworkSetupDockerImage    string
	PodInfraContainerImage     string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
}

type MachineDeployment struct {
	*controller.Controller
}

func NewMachineDeployment(config MachineDeploymentConfig) (*MachineDeployment, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	resourceSets, err := newMachineDeploymentResourceSets(config)
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
			// like operatorkit.giantswarm.io/aws-operator-machine-deployment-controller.
			Name: project.Name() + "-machine-deployment-controller",
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSMachineDeployment)
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

	c := &MachineDeployment{
		Controller: operatorkitController,
	}

	return c, nil
}

func newMachineDeploymentResourceSets(config MachineDeploymentConfig) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := machineDeploymentResourceSetConfig(config)

		resourceSet, err = newMachineDeploymentResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
