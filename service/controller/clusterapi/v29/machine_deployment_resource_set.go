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

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/credential"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/encryption"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/machinedeploymentsubnet"
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

	var machineDeploymentSubnetResource controller.Resource
	{
		c := machinedeploymentsubnet.Config{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		machineDeploymentSubnetResource, err = machinedeploymentsubnet.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		encryptionResource,
		machineDeploymentSubnetResource,
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
		var cr v1alpha1.Cluster
		{
			md, err := key.ToMachineDeployment(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			id := key.WorkerClusterID(md)

			m, err := config.CMAClient.ClusterV1alpha1().Clusters(md.Namespace).Get(id, metav1.GetOptions{})
			if err != nil {
				return nil, microerror.Mask(err)
			}

			cr = *m
		}

		var tenantClusterAWSClients aws.Clients
		{
			arn, err := credential.GetARN(config.K8sClient, cr)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			c := config.HostAWSConfig
			c.RoleARN = arn

			tenantClusterAWSClients, err = aws.NewClients(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		cc := controllercontext.Context{
			Client: controllercontext.ContextClient{
				ControlPlane: controllercontext.ContextClientControlPlane{
					AWS: config.ControlPlaneAWSClients,
				},
				TenantCluster: controllercontext.ContextClientTenantCluster{
					AWS: tenantClusterAWSClients,
				},
			},
		}
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
