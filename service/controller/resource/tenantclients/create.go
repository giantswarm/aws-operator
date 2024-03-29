package tenantclients

import (
	"context"

	"github.com/aws/amazon-vpc-cni-k8s/pkg/apis/crd/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v6/pkg/tenantcluster"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/aws-operator/v16/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v16/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var restConfig *rest.Config
	{
		restConfig, err = r.tenant.NewRestConfig(ctx, key.ClusterID(&cr), key.ClusterAPIEndpoint(cr))
		if tenantcluster.IsTimeout(err) {
			r.logger.Debugf(ctx, "timeout fetching certificates")
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger: r.logger,

			RestConfig: rest.CopyConfig(restConfig),
			SchemeBuilder: k8sclient.SchemeBuilder{
				// The Tenant Clients are used to connect to manage ENIConfig CRs within
				// the Tenant Cluster in order to properly configure the AWS CNI.
				// Therefore it is important to add its specific scheme builders.
				v1alpha1.AddToScheme,
				// The Tenant Clients are used to delete the ENIConfig CRD after the migration to Cilium.
				apiextensionsv1.AddToScheme,
			},
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			// On any error we want to handle the situation gracefully in order
			// to not block the whole reconciliation. Our former approach of
			// matching against specific errors was extremely brittle because
			// every now and then new errors popped up which we never knew or
			// handled before. This is extremely painful after fact in a
			// immutable infrastructure because it is super hard to fix once it
			// breaks after it is released.
			r.logger.Debugf(ctx, "tenant API not available yet", "stack", microerror.JSON(err))
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}
	}

	{
		cc.Client.TenantCluster.K8s = k8sClient
	}

	return nil
}
