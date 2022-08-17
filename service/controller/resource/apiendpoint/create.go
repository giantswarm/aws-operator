package apiendpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/pkg/label"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cluster := &apiv1beta1.Cluster{}
	{
		var clusters apiv1beta1.ClusterList

		var labelSelector client.MatchingLabels
		{
			labelSelector = make(map[string]string)
			labelSelector[label.Cluster] = key.ClusterID(&cr)
		}

		err = r.ctrlClient.List(ctx, &clusters, labelSelector)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(clusters.Items) == 0 {
			return microerror.Mask(notFoundError)
		} else if len(clusters.Items) > 1 {
			objName := fmt.Sprintf("%s.%s", clusters.Items[0].APIVersion, clusters.Items[0].Kind)
			return microerror.Maskf(tooManyResultsError, "got %d, expected 1 %s with label %s=%s", len(clusters.Items), objName, label.Cluster, key.ClusterID(&cr))
		}

		clusters.Items[0].DeepCopyInto(cluster)
	}

	{
		apiEndpoint := apiv1beta1.APIEndpoint{
			Host: key.ClusterAPIEndpoint(cr),
			Port: 443,
		}

		if cluster.Spec.ControlPlaneEndpoint.Host == apiEndpoint.Host && cluster.Spec.ControlPlaneEndpoint.Port == apiEndpoint.Port {
			r.logger.Debugf(ctx, "API endpoint already set")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		cluster.Spec.ControlPlaneEndpoint.Host = apiEndpoint.Host
		cluster.Spec.ControlPlaneEndpoint.Port = apiEndpoint.Port

	}

	{
		err = r.ctrlClient.Update(ctx, cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "API endpoint set")
	}

	return nil
}
