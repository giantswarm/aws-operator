package tcnpstatus

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "tcnpstatus"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		instanceTypesEqual := reflect.DeepEqual(cr.Status.Provider.Worker.InstanceTypes, cc.Status.TenantCluster.TCNP.Instances.InstanceTypes)
		numberInstancesEqual := cr.Status.Provider.Worker.SpotInstances == cc.Status.TenantCluster.TCNP.Instances.NumberOfSpotInstances

		if !instanceTypesEqual || !numberInstancesEqual {
			r.logger.LogCtx(ctx, "level", "debug", "message", "updating cr status")

			cr.Status.Provider.Worker.InstanceTypes = cc.Status.TenantCluster.TCNP.Instances.InstanceTypes
			cr.Status.Provider.Worker.SpotInstances = cc.Status.TenantCluster.TCNP.Instances.NumberOfSpotInstances

			_, err := r.g8sClient.InfrastructureV1alpha2().AWSMachineDeployments(cr.Namespace).UpdateStatus(&cr)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "updated cr status")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return nil
		}
	}

	return nil
}
