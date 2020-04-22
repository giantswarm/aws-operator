package controller

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/randomkeys"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/internal/adapter"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/internal/network"
)

type ClusterConfig struct {
	K8sClient        k8sclient.Interface
	Logger           micrologger.Logger
	NetworkAllocator network.Allocator

	AccessLogsExpiration       int
	AdvancedMonitoringEC2      bool
	APIWhitelist               ClusterConfigAPIWhitelist
	DeleteLoggingBucket        bool
	EncrypterBackend           string
	GuestAWSConfig             ClusterConfigAWSConfig
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	GuestSubnetMaskBits        int
	GuestUpdateEnabled         bool
	HostAWSConfig              ClusterConfigAWSConfig
	IgnitionPath               string
	ImagePullProgressDeadline  string
	IncludeTags                bool
	InstallationName           string
	IPAMNetworkRange           net.IPNet
	LabelSelector              ClusterConfigLabelSelector
	OIDC                       ClusterConfigOIDC
	PodInfraContainerImage     string
	ProjectName                string
	RegistryDomain             string
	Route53Enabled             bool
	RouteTables                string
	SSOPublicKey               string
	VaultAddress               string
}

type ClusterConfigAWSConfig struct {
	AccessKeyID       string
	AccessKeySecret   string
	AvailabilityZones []string
	Region            string
	SessionToken      string
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

// ClusterConfigAPIWhitelist defines guest cluster k8s API whitelisting types.
type ClusterConfigAPIWhitelist struct {
	Private ClusterConfigAPIWhitelistConfig
	Public  ClusterConfigAPIWhitelistConfig
}

// ClusterConfigAPIWhitelistConfig defines guest cluster k8s API whitelisting.
type ClusterConfigAPIWhitelistConfig struct {
	Enabled    bool
	SubnetList string
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	resourceSets, err := newClusterResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			NewRuntimeObjectFunc: func() pkgruntime.Object {
				return new(v1alpha1.AWSConfig)
			},

			Name: config.ProjectName,
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: clusterController,
	}

	return c, nil
}

func newClusterResourceSets(config ClusterConfig) ([]*controller.ResourceSet, error) {
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
		c := clusterResourceSetConfig{
			CertsSearcher:          certsSearcher,
			ControlPlaneAWSClients: controlPlaneAWSClients,
			HostAWSConfig: awsclient.Config{
				AccessKeyID:     config.HostAWSConfig.AccessKeyID,
				AccessKeySecret: config.HostAWSConfig.AccessKeySecret,
				Region:          config.HostAWSConfig.Region,
				SessionToken:    config.HostAWSConfig.SessionToken,
			},
			K8sClient:          config.K8sClient,
			Logger:             config.Logger,
			NetworkAllocator:   config.NetworkAllocator,
			RandomKeysSearcher: randomKeysSearcher,

			AccessLogsExpiration:  config.AccessLogsExpiration,
			AdvancedMonitoringEC2: config.AdvancedMonitoringEC2,
			APIWhitelist: adapter.APIWhitelist{
				Private: adapter.Whitelist{
					Enabled:    config.APIWhitelist.Private.Enabled,
					SubnetList: config.APIWhitelist.Private.SubnetList,
				},
				Public: adapter.Whitelist{
					Enabled:    config.APIWhitelist.Public.Enabled,
					SubnetList: config.APIWhitelist.Public.SubnetList,
				},
			},
			DeleteLoggingBucket:        config.DeleteLoggingBucket,
			EncrypterBackend:           config.EncrypterBackend,
			GuestAvailabilityZones:     config.GuestAWSConfig.AvailabilityZones,
			GuestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
			GuestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
			GuestSubnetMaskBits:        config.GuestSubnetMaskBits,
			PodInfraContainerImage:     config.PodInfraContainerImage,
			Route53Enabled:             config.Route53Enabled,
			IgnitionPath:               config.IgnitionPath,
			ImagePullProgressDeadline:  config.ImagePullProgressDeadline,
			IncludeTags:                config.IncludeTags,
			InstallationName:           config.InstallationName,
			IPAMNetworkRange:           config.IPAMNetworkRange,
			OIDC: cloudconfig.OIDCConfig{
				ClientID:      config.OIDC.ClientID,
				IssuerURL:     config.OIDC.IssuerURL,
				UsernameClaim: config.OIDC.UsernameClaim,
				GroupsClaim:   config.OIDC.GroupsClaim,
			},
			ProjectName:    config.ProjectName,
			RouteTables:    config.RouteTables,
			RegistryDomain: config.RegistryDomain,
			SSOPublicKey:   config.SSOPublicKey,
			VaultAddress:   config.VaultAddress,
		}

		resourceSet, err = newClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
