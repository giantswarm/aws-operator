package tcnp

import (
	"context"
	"fmt"
	"strconv"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Resource) disableClusterAutoscaler(ctx context.Context, awscluster infrastructurev1alpha3.AWSCluster, wcclient ctrlClient.Client) error {
	// Add or update an annotation to add 1 to the number of NPs blocking cluster-autoscaler
	_, err := r.updateAnnotation(ctx, awscluster, 1)
	if err != nil {
		return microerror.Mask(err)
	}

	// Scale cluster autoscaler to 0 replicas
	err = r.scaleAutoscaler(ctx, wcclient, 0)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) enableClusterAutoscaler(ctx context.Context, awscluster infrastructurev1alpha3.AWSCluster, wcclient ctrlClient.Client) error {
	// Add or update an annotation to decrease 1 to the number of NPs blocking cluster-autoscaler
	desired, err := r.updateAnnotation(ctx, awscluster, -1)
	if err != nil {
		return microerror.Mask(err)
	}

	if desired > 0 {
		r.logger.Debugf(ctx, "there are %d MPs pending upgrade. NOT enabling cluster autoscaler", desired)
		return nil
	}

	// Scale cluster autoscaler to 1 replica
	err = r.scaleAutoscaler(ctx, wcclient, 1)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) updateAnnotation(ctx context.Context, awscluster infrastructurev1alpha3.AWSCluster, diff int) (int32, error) {
	current := awscluster.Annotations[MDBlockingClusterAutoscalerCountAnnotation]
	desired := 1
	if current != "" {
		val, err := strconv.Atoi(current)
		if err != nil {
			return -1, microerror.Mask(err)
		}

		desired = val + diff
	}

	r.logger.Debugf(ctx, "setting %s annotation to %d", MDBlockingClusterAutoscalerCountAnnotation, desired)

	if awscluster.Annotations == nil {
		awscluster.Annotations = make(map[string]string)
	}

	if desired > 0 {
		awscluster.Annotations[MDBlockingClusterAutoscalerCountAnnotation] = fmt.Sprint(desired)
	} else {
		delete(awscluster.Annotations, MDBlockingClusterAutoscalerCountAnnotation)
	}

	err := r.k8sClient.CtrlClient().Update(ctx, &awscluster)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "set %s annotation to %d", MDBlockingClusterAutoscalerCountAnnotation, desired)

	return int32(desired), nil
}

func (r *Resource) scaleAutoscaler(ctx context.Context, wcclient ctrlClient.Client, replicas int32) error {
	r.logger.Debugf(ctx, "scaling cluster autoscaler to %d replicas", replicas)

	autoscaler := v1.Deployment{}
	err := wcclient.Get(ctx, ctrlClient.ObjectKey{Name: "cluster-autoscaler", Namespace: "kube-system"}, &autoscaler)
	if err != nil {
		return microerror.Mask(err)
	}

	autoscaler.Spec.Replicas = &replicas

	err = wcclient.Update(ctx, &autoscaler)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "scaled cluster autoscaler to %d replicas", replicas)

	return nil
}
