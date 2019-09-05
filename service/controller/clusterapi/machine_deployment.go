package clusterapi

import (
	"net"

	clusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	v29 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v29"
	v30 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v30"
	v30cloudconfig "github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/locker"
)

type MachineDeploymentConfig struct {
	CMAClient    clientset.Interface
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Locker       locker.Interface
	Logger       micrologger.Logger

	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DeleteLoggingBucket        bool
	DockerDaemonCIDR           string
	EncrypterBackend           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	HostAWSConfig              aws.Config
	IgnitionPath               string
	ImagePullProgressDeadline  string
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	LabelSelector              MachineDeploymentConfigLabelSelector
	NetworkSetupDockerImage    string
	OIDC                       ClusterConfigOIDC
	PodInfraContainerImage     string
	ProjectName                string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
	VPCPeerID                  string
}

type MachineDeploymentConfigLabelSelector struct {
	Enabled          bool
	OverridenVersion string
}

type MachineDeployment struct {
	*controller.Controller
}

func NewMachineDeployment(config MachineDeploymentConfig) (*MachineDeployment, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
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
			Watcher: config.CMAClient.ClusterV1alpha1().MachineDeployments(corev1.NamespaceAll),

			ListOptions: metav1.ListOptions{
				LabelSelector: key.VersionLabelSelector(config.LabelSelector.Enabled, config.LabelSelector.OverridenVersion),
			},
			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets, err := newMachineDeploymentResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          clusterv1alpha1.NewMachineDeploymentCRD(),
			CRDClient:    crdClient,
			Informer:     newInformer,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			RESTClient:   config.CMAClient.ClusterV1alpha1().RESTClient(),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/aws-operator-machine-deployment-controller.
			Name: config.ProjectName + "-machine-deployment-controller",
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

	var controlPlaneAWSClients aws.Clients
	{
		c := aws.Config{
			AccessKeyID:     config.HostAWSConfig.AccessKeyID,
			AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
			Region:          config.HostAWSConfig.Region,
			SessionToken:    config.HostAWSConfig.SessionToken,
		}

		controlPlaneAWSClients, err = aws.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v29ResourceSet *controller.ResourceSet
	{
		c := v29.MachineDeploymentResourceSetConfig{
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			K8sClient:              config.K8sClient,
			Locker:                 config.Locker,
			Logger:                 config.Logger,

			EncrypterBackend:           config.EncrypterBackend,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			HostAWSConfig:              config.HostAWSConfig,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			ProjectName:                config.ProjectName,
			Route53Enabled:             config.Route53Enabled,
			RouteTables:                config.RouteTables,
			VaultAddress:               config.VaultAddress,
			VPCPeerID:                  config.VPCPeerID,
		}

		v29ResourceSet, err = v29.NewMachineDeploymentResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v30ResourceSet *controller.ResourceSet
	{
		c := v30.MachineDeploymentResourceSetConfig{
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			K8sClient:              config.K8sClient,
			Locker:                 config.Locker,
			Logger:                 config.Logger,

			CalicoCIDR:                 config.CalicoCIDR,
			CalicoMTU:                  config.CalicoMTU,
			CalicoSubnet:               config.CalicoSubnet,
			ClusterIPRange:             config.ClusterIPRange,
			DockerDaemonCIDR:           config.DockerDaemonCIDR,
			EncrypterBackend:           config.EncrypterBackend,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			HostAWSConfig:              config.HostAWSConfig,
			IgnitionPath:               config.IgnitionPath,
			ImagePullProgressDeadline:  config.ImagePullProgressDeadline,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			NetworkSetupDockerImage:    config.NetworkSetupDockerImage,
			OIDC: v30cloudconfig.ConfigOIDC{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			PodInfraContainerImage: config.PodInfraContainerImage,
			ProjectName:            config.ProjectName,
			RegistryDomain:         config.RegistryDomain,
			Route53Enabled:         config.Route53Enabled,
			RouteTables:            config.RouteTables,
			SSHUserList:            config.SSHUserList,
			SSOPublicKey:           config.SSOPublicKey,
			VaultAddress:           config.VaultAddress,
			VPCPeerID:              config.VPCPeerID,
		}

		v30ResourceSet, err = v30.NewMachineDeploymentResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		v29ResourceSet,
		v30ResourceSet,
	}

	return resourceSets, nil
}
