package controller

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/cpvpc"
	"github.com/giantswarm/aws-operator/service/controller/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/resource/s3object"
	"github.com/giantswarm/aws-operator/service/controller/resource/snapshotid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpazs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpn"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpnencryption"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpnoutputs"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsecuritygroups"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpsubnets"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcid"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpvpcpcx"
)

type controlPlaneResourceSetConfig struct {
	G8sClient          versioned.Interface
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	CertsSearcher      certs.Interface
	RandomKeysSearcher randomkeys.Interface

	APIWhitelist              tccpn.APIWhitelist
	CalicoCIDR                int
	CalicoMTU                 int
	CalicoSubnet              string
	ClusterDomain             string
	ClusterIPRange            string
	DockerDaemonCIDR          string
	IgnitionPath              string
	ImagePullProgressDeadline string
	InstallationName          string
	HostAWSConfig             aws.Config
	NetworkSetupDockerImage   string
	PodInfraContainerImage    string
	RegistryDomain            string
	SSHUserList               string
	SSOPublicKey              string
	VaultAddress              string
}

func (c controlPlaneResourceSetConfig) GetInstallationName() string {
	return c.InstallationName
}

func (c controlPlaneResourceSetConfig) GetLogger() micrologger.Logger {
	return c.Logger
}

func (c controlPlaneResourceSetConfig) GetVaultAddress() string {
	return c.VaultAddress
}

func newControlPlaneResourceSet(config controlPlaneResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var awsClientResource resource.Interface
	{
		c := awsclient.Config{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),

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
		encrypterObject, err = newEncrypterObject(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnCloudConfig *cloudconfig.TCCPN
	{
		c := cloudconfig.TCCPNConfig{
			Config: cloudconfig.Config{
				Encrypter: encrypterObject,
				Logger:    config.Logger,

				CalicoCIDR:                config.CalicoCIDR,
				CalicoMTU:                 config.CalicoMTU,
				CalicoSubnet:              config.CalicoSubnet,
				ClusterDomain:             config.ClusterDomain,
				ClusterIPRange:            config.ClusterIPRange,
				DockerDaemonCIDR:          config.DockerDaemonCIDR,
				IgnitionPath:              config.IgnitionPath,
				ImagePullProgressDeadline: config.ImagePullProgressDeadline,
				NetworkSetupDockerImage:   config.NetworkSetupDockerImage,
				PodInfraContainerImage:    config.PodInfraContainerImage,
				RegistryDomain:            config.RegistryDomain,
				SSHUserList:               config.SSHUserList,
				SSOPublicKey:              config.SSOPublicKey,
			},
			G8sClient: config.G8sClient,
		}

		tccpnCloudConfig, err = cloudconfig.NewTCCPN(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnChangeDetection *changedetection.TCCPN
	{
		c := changedetection.TCCPNConfig{
			Logger: config.Logger,
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),
		}

		tccpVPCPCXResource, err = tccpvpcpcx.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpnEncryptionResource resource.Interface
	{
		c := tccpnencryption.Config{
			G8sClient: config.G8sClient,
			Encrypter: encrypterObject,
			Logger:    config.Logger,
		}

		tccpnEncryptionResource, err = tccpnencryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var s3ObjectResource resource.Interface
	{
		c := s3object.Config{
			CertsSearcher:      config.CertsSearcher,
			CloudConfig:        tccpnCloudConfig,
			LabelsFunc:         key.KubeletLabelsTCCPN,
			Logger:             config.Logger,
			G8sClient:          config.G8sClient,
			PathFunc:           key.S3ObjectPathTCCPN,
			RandomKeysSearcher: config.RandomKeysSearcher,
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
			G8sClient:     config.G8sClient,
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),

			CIDRBlockAWSCNI: fmt.Sprintf("%s/%d", config.CalicoSubnet, config.CalicoCIDR),
		}

		tccpAZsResource, err = tccpazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tccpSecurityGroupsResource resource.Interface
	{
		c := tccpsecuritygroups.Config{
			Logger:        config.Logger,
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),
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
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			APIWhitelist:     config.APIWhitelist,
			Detection:        tccpnChangeDetection,
			InstallationName: config.InstallationName,
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
			ToClusterFunc: newControlPlaneToClusterFunc(config.G8sClient),
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
		tccpnOutputsResource,
		tccpnEncryptionResource,
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

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToControlPlane(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == project.Version() {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func newControlPlaneToClusterFunc(g8sClient versioned.Interface) func(obj interface{}) (infrastructurev1alpha2.AWSCluster, error) {
	return func(obj interface{}) (infrastructurev1alpha2.AWSCluster, error) {
		cr, err := key.ToControlPlane(obj)
		if err != nil {
			return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
		}

		m, err := g8sClient.InfrastructureV1alpha2().AWSClusters(cr.Namespace).Get(key.ClusterID(&cr), metav1.GetOptions{})
		if err != nil {
			return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
