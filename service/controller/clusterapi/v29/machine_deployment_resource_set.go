package v29

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/clusterazs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/encryption"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/ipam"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/machinedeploymentazs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/region"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tcnp"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/vpccidr"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/vpcid"
)

func NewMachineDeploymentResourceSet(config MachineDeploymentResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var encrypterObject encrypter.Interface
	{
		encrypterObject, err = newEncrypterObject(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentChecker *ipam.MachineDeploymentChecker
	{
		c := ipam.MachineDeploymentCheckerConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentChecker, err = ipam.NewMachineDeploymentChecker(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var subnetCollector *ipam.SubnetCollector
	{
		c := ipam.SubnetCollectorConfig{
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			NetworkRange: config.IPAMNetworkRange,
		}

		subnetCollector, err = ipam.NewSubnetCollector(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentPersister *ipam.MachineDeploymentPersister
	{
		c := ipam.MachineDeploymentPersisterConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentPersister, err = ipam.NewMachineDeploymentPersister(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsClientResource controller.Resource
	{
		c := awsclient.Config{
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterAZsResource controller.Resource
	{
		c := clusterazs.Config{
			CMAClient:     config.CMAClient,
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		clusterAZsResource, err = clusterazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionResource controller.Resource
	{
		c := encryption.Config{
			Encrypter:     encrypterObject,
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		encryptionResource, err = encryption.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ipamResource controller.Resource
	{
		c := ipam.Config{
			Checker:   machineDeploymentChecker,
			Collector: subnetCollector,
			Locker:    config.Locker,
			Logger:    config.Logger,
			Persister: machineDeploymentPersister,

			AllocatedSubnetMaskBits: config.GuestSubnetMaskBits,
			NetworkRange:            config.IPAMNetworkRange,
			PrivateSubnetMaskBits:   config.GuestPrivateSubnetMaskBits,
			PublicSubnetMaskBits:    config.GuestPublicSubnetMaskBits,
		}

		ipamResource, err = ipam.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machineDeploymentAZsResource controller.Resource
	{
		c := machinedeploymentazs.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentAZsResource, err = machinedeploymentazs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var regionResource controller.Resource
	{
		c := region.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		regionResource, err = region.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcnpResource controller.Resource
	{
		c := tcnp.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,

			InstallationName: config.InstallationName,
		}

		tcnpResource, err = tcnp.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vpcCIDRResource controller.Resource
	{
		c := vpccidr.Config{
			Logger: config.Logger,

			VPCPeerID: config.VPCPeerID,
		}

		vpcCIDRResource, err = vpccidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vpcidResource controller.Resource
	{
		c := vpcid.Config{
			CMAClient:     config.CMAClient,
			Logger:        config.Logger,
			ToClusterFunc: newMachineDeploymentToClusterFunc(config.CMAClient),
		}

		vpcidResource, err = vpcid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		awsClientResource,
		vpcidResource,
		vpcCIDRResource,
		regionResource,
		encryptionResource,
		ipamResource,
		clusterAZsResource,
		machineDeploymentAZsResource,
		tcnpResource,
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
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		cc := controllercontext.Context{}
		ctx = controllercontext.NewContext(ctx, cc)

		return ctx, nil
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

func newMachineDeploymentToClusterFunc(cmaClient clientset.Interface) func(obj interface{}) (v1alpha1.Cluster, error) {
	return func(obj interface{}) (v1alpha1.Cluster, error) {
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return v1alpha1.Cluster{}, microerror.Mask(err)
		}

		m, err := cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Get(key.ClusterID(&cr), metav1.GetOptions{})
		if err != nil {
			return v1alpha1.Cluster{}, microerror.Mask(err)
		}

		return *m, nil
	}
}
