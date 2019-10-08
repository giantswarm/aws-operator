package clusterapi

import (
	"net"

	clusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/randomkeys"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
	v30 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v30"
	v30adapter "github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/adapter"
	v30cloudconfig "github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/cloudconfig"
	v31 "github.com/giantswarm/aws-operator/service/controller/clusterapi/v31"
	v31adapter "github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/adapter"
	v31cloudconfig "github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/locker"
)

type ClusterConfig struct {
	CMAClient    clientset.Interface
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Locker       locker.Interface
	Logger       micrologger.Logger

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               FrameworkConfigAPIWhitelist
	CalicoCIDR                 int
	CalicoMTU                  int
	CalicoSubnet               string
	ClusterIPRange             string
	DeleteLoggingBucket        bool
	DockerDaemonCIDR           string
	EncrypterBackend           string
	GuestAvailabilityZones     []string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	GuestUpdateEnabled         bool
	HostAWSConfig              aws.Config
	IgnitionPath               string
	ImagePullProgressDeadline  string
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	LabelSelector              ClusterConfigLabelSelector
	NetworkSetupDockerImage    string
	OIDC                       ClusterConfigOIDC
	PodInfraContainerImage     string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSHUserList                string
	SSOPublicKey               string
	VaultAddress               string
	VPCPeerID                  string
}

type ClusterConfigLabelSelector struct {
	Enabled          bool
	OverridenVersion string
}

// ClusterConfigOIDC represents the configuration of the OIDC authorization
// provider.
type ClusterConfigOIDC struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

// FrameworkConfigAPIWhitelist defines guest cluster k8s API whitelisting types.
type FrameworkConfigAPIWhitelist struct {
	Private FrameworkConfigAPIWhitelistConfig
	Public  FrameworkConfigAPIWhitelistConfig
}

// FrameworkConfigAPIWhitelistConfig defines guest cluster k8s API whitelisting.
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
			Watcher: config.CMAClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll),

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

	resourceSets, err := newClusterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          clusterv1alpha1.NewClusterCRD(),
			CRDClient:    crdClient,
			Informer:     newInformer,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			RESTClient:   config.CMAClient.ClusterV1alpha1().RESTClient(),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/aws-operator-cluster-controller.
			Name: project.Name() + "-cluster-controller",
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

	var certsSearcher *certs.Searcher
	{
		c := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		certsSearcher, err = certs.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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

	var randomKeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		randomKeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV30 *controller.ResourceSet
	{
		c := v30.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			HostAWSConfig:          config.HostAWSConfig,
			K8sClient:              config.K8sClient,
			Locker:                 config.Locker,
			Logger:                 config.Logger,
			RandomKeysSearcher:     randomKeysSearcher,

			AccessLogsExpiration:  config.AccessLogsExpiration,
			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			APIWhitelist: v30adapter.APIWhitelist{
				Private: v30adapter.Whitelist{
					Enabled:    config.APIWhitelist.Private.Enabled,
					SubnetList: config.APIWhitelist.Private.SubnetList,
				},
				Public: v30adapter.Whitelist{
					Enabled:    config.APIWhitelist.Public.Enabled,
					SubnetList: config.APIWhitelist.Public.SubnetList,
				},
			},
			CalicoCIDR:                 config.CalicoCIDR,
			CalicoMTU:                  config.CalicoMTU,
			CalicoSubnet:               config.CalicoSubnet,
			ClusterIPRange:             config.ClusterIPRange,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			DockerDaemonCIDR:           config.DockerDaemonCIDR,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			IgnitionPath:               config.IgnitionPath,
			ImagePullProgressDeadline:  config.ImagePullProgressDeadline,
			IncludeTags:                config.IncludeTags,
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
			RegistryDomain:         config.RegistryDomain,
			Route53Enabled:         config.Route53Enabled,
			RouteTables:            config.RouteTables,
			SSHUserList:            config.SSHUserList,
			SSOPublicKey:           config.SSOPublicKey,
			VaultAddress:           config.VaultAddress,
			VPCPeerID:              config.VPCPeerID,
		}

		resourceSetV30, err = v30.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV31 *controller.ResourceSet
	{
		c := v31.ClusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			CMAClient:              config.CMAClient,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			G8sClient:              config.G8sClient,
			HostAWSConfig:          config.HostAWSConfig,
			K8sClient:              config.K8sClient,
			Locker:                 config.Locker,
			Logger:                 config.Logger,
			RandomKeysSearcher:     randomKeysSearcher,

			AccessLogsExpiration:  config.AccessLogsExpiration,
			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			APIWhitelist: v31adapter.APIWhitelist{
				Private: v31adapter.Whitelist{
					Enabled:    config.APIWhitelist.Private.Enabled,
					SubnetList: config.APIWhitelist.Private.SubnetList,
				},
				Public: v31adapter.Whitelist{
					Enabled:    config.APIWhitelist.Public.Enabled,
					SubnetList: config.APIWhitelist.Public.SubnetList,
				},
			},
			CalicoCIDR:                 config.CalicoCIDR,
			CalicoMTU:                  config.CalicoMTU,
			CalicoSubnet:               config.CalicoSubnet,
			ClusterIPRange:             config.ClusterIPRange,
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			DockerDaemonCIDR:           config.DockerDaemonCIDR,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			IgnitionPath:               config.IgnitionPath,
			ImagePullProgressDeadline:  config.ImagePullProgressDeadline,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			NetworkSetupDockerImage:    config.NetworkSetupDockerImage,
			OIDC: v31cloudconfig.ConfigOIDC{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			PodInfraContainerImage: config.PodInfraContainerImage,
			RegistryDomain:         config.RegistryDomain,
			Route53Enabled:         config.Route53Enabled,
			RouteTables:            config.RouteTables,
			SSHUserList:            config.SSHUserList,
			SSOPublicKey:           config.SSOPublicKey,
			VaultAddress:           config.VaultAddress,
			VPCPeerID:              config.VPCPeerID,
		}

		resourceSetV31, err = v31.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSetV30,
		resourceSetV31,
	}

	return resourceSets, nil
}
