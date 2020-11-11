package controller

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpvpc"
	"github.com/giantswarm/aws-operator/service/controller/resource/encryptionsearcher"
	"github.com/giantswarm/aws-operator/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/resource/snapshotid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpn"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpnoutputs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcpcx"
	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/images"
	event "github.com/giantswarm/aws-operator/service/internal/recorder"
	"github.com/giantswarm/aws-operator/service/internal/releases"
)

type ControlPlaneConfig struct {
	CertsSearcher      certs.Interface
	CloudTags          cloudtags.Interface
	Event              event.Interface
	HAMaster           hamaster.Interface
	Images             images.Interface
	K8sClient          k8sclient.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	CalicoCIDR                int
	CalicoMTU                 int
	CalicoSubnet              string
	ClusterDomain             string
	ClusterIPRange            string
	DockerDaemonCIDR          string
	DockerhubToken            string
	ExternalSNAT              bool
	HostAWSConfig             aws.Config
	IgnitionPath              string
	ImagePullProgressDeadline string
	InstallationName          string
	NetworkSetupDockerImage   string
	PodInfraContainerImage    string
	RegistryDomain            string
	RegistryMirrors           []string
	Route53Enabled            bool
	SSHUserList               string
	SSOPublicKey              string
}

type ControlPlane struct {
	*controller.Controller
}

func NewControlPlane(config ControlPlaneConfig) (*ControlPlane, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	var resources []resource.Interface
	{
		resources, err = newControlPlaneResources(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			InitCtx: func(ctx context.Context, obj interface{}) (context.Context, error) {
				return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.AWSControlPlane)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-control-plane-controller.
			Name: project.Name() + "-control-plane-controller",
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

func newControlPlaneResources(config ControlPlaneConfig) ([]resource.Interface, error) {
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

	var awsClientResource resource.Interface
	{
		c := awsclient.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var accountIDResource resource.Interface
	{
		c := accountid.Config{
			Logger: config.Logger,
		}

		accountIDResource, err = accountid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encrypterObject encrypter.Interface
	{
		c := &kms.EncrypterConfig{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		encrypterObject, err = kms.NewEncrypter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnCloudConfig *cloudconfig.TCCPN
	{
		c := cloudconfig.TCCPNConfig{
			Config: cloudconfig.Config{
				CertsSearcher:      certsSearcher,
				Encrypter:          encrypterObject,
				Event:              config.Event,
				HAMaster:           config.HAMaster,
				Images:             config.Images,
				K8sClient:          config.K8sClient,
				Logger:             config.Logger,
				RandomKeysSearcher: randomKeysSearcher,

				CalicoCIDR:                config.CalicoCIDR,
				CalicoMTU:                 config.CalicoMTU,
				CalicoSubnet:              config.CalicoSubnet,
				ClusterDomain:             config.ClusterDomain,
				ClusterIPRange:            config.ClusterIPRange,
				DockerDaemonCIDR:          config.DockerDaemonCIDR,
				DockerhubToken:            config.DockerhubToken,
				ExternalSNAT:              config.ExternalSNAT,
				IgnitionPath:              config.IgnitionPath,
				ImagePullProgressDeadline: config.ImagePullProgressDeadline,
				NetworkSetupDockerImage:   config.NetworkSetupDockerImage,
				PodInfraContainerImage:    config.PodInfraContainerImage,
				RegistryDomain:            config.RegistryDomain,
				RegistryMirrors:           config.RegistryMirrors,
				SSHUserList:               config.SSHUserList,
				SSOPublicKey:              config.SSOPublicKey,
			},
		}

		tccpnCloudConfig, err = cloudconfig.NewTCCPN(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var rel releases.Interface
	{
		c := releases.Config{
			K8sClient: config.K8sClient,
		}

		rel, err = releases.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnChangeDetection *changedetection.TCCPN
	{
		c := changedetection.TCCPNConfig{
			CloudTags: config.CloudTags,
			HAMaster:  config.HAMaster,
			Logger:    config.Logger,
			Event:     config.Event,
			Releases:  rel,
		}

		tccpnChangeDetection, err = changedetection.NewTCCPN(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var regionResource resource.Interface
	{
		c := region.Config{
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		regionResource, err = region.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpVPCPCXResource resource.Interface
	{
		c := tccpvpcpcx.Config{
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		tccpVPCPCXResource, err = tccpvpcpcx.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionSearcherResource resource.Interface
	{
		c := encryptionsearcher.Config{
			G8sClient:     config.K8sClient.G8sClient(),
			Encrypter:     encrypterObject,
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		encryptionSearcherResource, err = encryptionsearcher.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3ObjectResource resource.Interface
	{
		c := s3object.Config{
			CloudConfig: tccpnCloudConfig,
			Logger:      config.Logger,
		}

		ops, err := s3object.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		s3ObjectResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var snapshotIDResource resource.Interface
	{
		c := snapshotid.Config{
			Logger: config.Logger,
		}

		snapshotIDResource, err = snapshotid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpAZsResource resource.Interface
	{
		c := tccpazs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			CIDRBlockAWSCNI: fmt.Sprintf("%s/%d", config.CalicoSubnet, config.CalicoCIDR),
		}

		tccpAZsResource, err = tccpazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpOutputsResource resource.Interface
	{
		c := tccpoutputs.Config{
			Logger: config.Logger,

			Route53Enabled: config.Route53Enabled,
			ToClusterFunc:  newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		tccpOutputsResource, err = tccpoutputs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSecurityGroupsResource resource.Interface
	{
		c := tccpsecuritygroups.Config{
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		tccpSecurityGroupsResource, err = tccpsecuritygroups.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSubnetsResource resource.Interface
	{
		c := tccpsubnets.Config{
			Logger: config.Logger,
		}

		tccpSubnetsResource, err = tccpsubnets.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnResource resource.Interface
	{
		c := tccpn.Config{
			CloudTags: config.CloudTags,
			Detection: tccpnChangeDetection,
			Event:     config.Event,
			HAMaster:  config.HAMaster,
			Images:    config.Images,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			InstallationName: config.InstallationName,
			Route53Enabled:   config.Route53Enabled,
		}

		tccpnResource, err = tccpn.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnOutputsResource resource.Interface
	{
		c := tccpnoutputs.Config{
			Logger: config.Logger,
		}

		tccpnOutputsResource, err = tccpnoutputs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpVPCResource resource.Interface
	{
		c := cpvpc.Config{
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		cpVPCResource, err = cpvpc.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpVPCIDResource resource.Interface
	{
		c := tccpvpcid.Config{
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.G8sClient()),
		}

		tccpVPCIDResource, err = tccpvpcid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// All these resources only fetch information from remote APIs and put them
		// into the controller context.
		awsClientResource,
		accountIDResource,
		encryptionSearcherResource,
		tccpOutputsResource,
		tccpnOutputsResource,
		snapshotIDResource,
		tccpAZsResource,
		tccpSecurityGroupsResource,
		tccpVPCIDResource,
		tccpVPCPCXResource,
		tccpSubnetsResource,
		cpVPCResource,
		regionResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		s3ObjectResource,
		tccpnResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}

func newControlPlaneToClusterFunc(g8sClient versioned.Interface) func(ctx context.Context, obj interface{}) (infrastructurev1alpha2.AWSCluster, error) {
	return func(ctx context.Context, obj interface{}) (infrastructurev1alpha2.AWSCluster, error) {
		cr, err := key.ToControlPlane(obj)
		if err != nil {
			return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
		}

		m, err := g8sClient.InfrastructureV1alpha2().AWSClusters(cr.Namespace).Get(ctx, key.ClusterID(&cr), metav1.GetOptions{})
		if err != nil {
			return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
