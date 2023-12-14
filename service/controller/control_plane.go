package controller

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v4/pkg/certs"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v8/pkg/controller"
	"github.com/giantswarm/operatorkit/v8/pkg/resource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys/v3"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/aws-operator/v15/client/aws"
	"github.com/giantswarm/aws-operator/v15/pkg/label"
	"github.com/giantswarm/aws-operator/v15/pkg/project"
	"github.com/giantswarm/aws-operator/v15/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v15/service/controller/key"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/cleanuptccpniamroles"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/cpvpc"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/snapshotid"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpn"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpnoutputs"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpoutputs"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/v15/service/controller/resource/tccpvpcpcx"
	"github.com/giantswarm/aws-operator/v15/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v15/service/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/v15/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v15/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v15/service/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/v15/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/v15/service/internal/images"
	event "github.com/giantswarm/aws-operator/v15/service/internal/recorder"
	"github.com/giantswarm/aws-operator/v15/service/internal/releases"

	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
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

	CalicoCIDR              int
	CalicoMTU               int
	CalicoSubnet            string
	ClusterDomain           string
	ClusterIPRange          string
	DockerDaemonCIDR        string
	DockerhubToken          string
	ExternalSNAT            bool
	HostAWSConfig           aws.Config
	IgnitionPath            string
	InstallationName        string
	NetworkSetupDockerImage string
	PodInfraContainerImage  string
	RegistryDomain          string
	RegistryMirrors         []string
	Route53Enabled          bool
	SSHUserList             string
	SSOPublicKey            string
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
			NewRuntimeObjectFunc: func() ctrlClient.Object {
				return new(infrastructurev1alpha3.AWSControlPlane)
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.CtrlClient()),

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

	var cleanupIAMRolesResource resource.Interface
	{
		c := cleanuptccpniamroles.Config{
			Logger: config.Logger,
		}

		cleanupIAMRolesResource, err = cleanuptccpniamroles.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cloudtagObject cloudtags.Interface
	{
		c := cloudtags.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		cloudtagObject, err = cloudtags.New(c)
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

				CalicoCIDR:              config.CalicoCIDR,
				CalicoMTU:               config.CalicoMTU,
				CalicoSubnet:            config.CalicoSubnet,
				ClusterDomain:           config.ClusterDomain,
				ClusterIPRange:          config.ClusterIPRange,
				DockerDaemonCIDR:        config.DockerDaemonCIDR,
				DockerhubToken:          config.DockerhubToken,
				ExternalSNAT:            config.ExternalSNAT,
				IgnitionPath:            config.IgnitionPath,
				NetworkSetupDockerImage: config.NetworkSetupDockerImage,
				PodInfraContainerImage:  config.PodInfraContainerImage,
				RegistryDomain:          config.RegistryDomain,
				RegistryMirrors:         config.RegistryMirrors,
				SSHUserList:             config.SSHUserList,
				SSOPublicKey:            config.SSOPublicKey,
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
			Event:    config.Event,
			HAMaster: config.HAMaster,
			Logger:   config.Logger,
			Releases: rel,
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.CtrlClient()),
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tccpVPCPCXResource, err = tccpvpcpcx.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3ObjectResource resource.Interface
	{
		c := s3object.Config{
			CloudConfig: tccpnCloudConfig,
			Encrypter:   encrypterObject,
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
			ToClusterFunc:  newControlPlaneToClusterFunc(config.K8sClient.CtrlClient()),
		}

		tccpOutputsResource, err = tccpoutputs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSecurityGroupsResource resource.Interface
	{
		c := tccpsecuritygroups.Config{
			CtrlClient:    config.K8sClient.CtrlClient(),
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.CtrlClient()),
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
			CloudTags: cloudtagObject,
			Detection: tccpnChangeDetection,
			Encrypter: encrypterObject,
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.K8sClient.CtrlClient()),
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

		// All these resources implement cleanup functionality only being executed
		// on delete events.
		cleanupIAMRolesResource,
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

func newControlPlaneToClusterFunc(ctrlClient ctrlClient.Client) func(ctx context.Context, obj interface{}) (infrastructurev1alpha3.AWSCluster, error) {
	return func(ctx context.Context, obj interface{}) (infrastructurev1alpha3.AWSCluster, error) {
		cr, err := key.ToControlPlane(obj)
		if err != nil {
			return infrastructurev1alpha3.AWSCluster{}, microerror.Mask(err)
		}

		m := &infrastructurev1alpha3.AWSCluster{}
		err = ctrlClient.Get(ctx, client.ObjectKey{Name: key.ClusterID(&cr), Namespace: cr.Namespace}, m)
		if err != nil {
			return infrastructurev1alpha3.AWSCluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
