package apiendpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var cluster apiv1alpha2.Cluster
	{
		var clusters apiv1alpha2.ClusterList

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

		clusters.Items[0].DeepCopyInto(&cluster)
	}

	{
		apiEndpoint := apiv1alpha2.APIEndpoint{
			Host: key.ClusterAPIEndpoint(cr),
			Port: 443,
		}

		for _, ep := range cluster.Status.APIEndpoints {
			if ep.Host == apiEndpoint.Host && ep.Port == apiEndpoint.Port {
				r.logger.LogCtx(ctx, "level", "debug", "message", "API endpoint already set")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				return nil
			}
		}

		cluster.Status.APIEndpoints = append(cluster.Status.APIEndpoints, apiEndpoint)
	}

	{
		err = r.ctrlClient.Status().Update(ctx, &cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "API endpoint set")
	}

	return nil
}
