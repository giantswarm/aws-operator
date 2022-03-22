package controller

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/accountid"
	"github.com/giantswarm/aws-operator/service/controller/resource/apiendpoint"
	"github.com/giantswarm/aws-operator/service/controller/resource/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/aws-operator/service/controller/resource/terminateunhealthynode"
	"github.com/giantswarm/aws-operator/service/internal/locker"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type TerminateUnhealthyNodeConfig struct {
	K8sClient k8sclient.Interface
	Locker    locker.Interface
	Logger    micrologger.Logger

	HostAWSConfig aws.Config
}

type TerminateUnhealthyNode struct {
	*controller.Controller
}

func NewTerminateUnhealthyNode(config TerminateUnhealthyNodeConfig) (*TerminateUnhealthyNode, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var resources []resource.Interface
	{
		resources, err = newTerminateUnhealthyNodeResources(config)
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
				return new(infrastructurev1alpha3.AWSCluster)
			},
			Resources:    resources,
			ResyncPeriod: key.TerminateUnhealthyNodeResyncPeriod,

			// Name is used to compute finalizer names. This results in something
			// like operatorkit.giantswarm.io/aws-operator-cluster-controller.
			Name: project.Name() + "-terminate-unhealthy-node-controller",
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &TerminateUnhealthyNode{
		Controller: operatorkitController,
	}

	return c, nil
}

func newTerminateUnhealthyNodeResources(config TerminateUnhealthyNodeConfig) ([]resource.Interface, error) {
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

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: certsSearcher,
			Logger:        config.Logger,
			CertID:        certs.AWSOperatorAPICert,
		}

		tenantCluster, err = tenantcluster.New(c)
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

	var apiEndpointResource resource.Interface
	{
		c := apiendpoint.Config{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		apiEndpointResource, err = apiendpoint.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var awsClientResource resource.Interface
	{
		c := awsclient.Config{
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			ToClusterFunc: key.ToCluster,

			CPAWSConfig: config.HostAWSConfig,
		}

		awsClientResource, err = awsclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger: config.Logger,
			Tenant: tenantCluster,

			ToClusterFunc: key.ToCluster,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var terminateUnhealthyNodeResource resource.Interface
	{
		c := terminateunhealthynode.Config{
			Logger: config.Logger,
		}

		terminateUnhealthyNodeResource, err = terminateunhealthynode.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// All these resources only fetch information from remote APIs and put them
		// into the controller context.
		awsClientResource,
		accountIDResource,
		tenantClientsResource,

		// All these resources implement certain business logic and operate based on
		// the information given in the controller context.
		apiEndpointResource,
		terminateUnhealthyNodeResource,
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
