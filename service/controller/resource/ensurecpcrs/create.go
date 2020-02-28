package ensurecpcrs

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring %#q CR exists", fmt.Sprintf("%T", infrastructurev1alpha2.AWSControlPlane{})))

		exists, err := r.awsControlPlaneCRExists(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if !exists {
			err := r.createAWSControlPlaneCR(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured %#q CR exists", fmt.Sprintf("%T", infrastructurev1alpha2.AWSControlPlane{})))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring %#q CR exists", fmt.Sprintf("%T", infrastructurev1alpha2.G8sControlPlane{})))

		exists, err := r.g8sControlPlaneCRExists(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if !exists {
			err := r.createG8sControlPlaneCR(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured %#q CR exists", fmt.Sprintf("%T", infrastructurev1alpha2.G8sControlPlane{})))
	}

	return nil
}

func (r *Resource) createAWSControlPlaneCR(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	return nil
}

func (r *Resource) createG8sControlPlaneCR(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	return nil
}

func (r *Resource) awsControlPlaneCRExists(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (bool, error) {
	o := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.Cluster, key.ClusterID(&cr)),
	}

	list, err := r.k8sClient.G8sClient().InfrastructureV1alpha2().AWSControlPlanes(cr.GetNamespace()).List(o)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// We simply list Control Plane CRs by filtering for the cluster ID label.
	// This may lead to one or more results as the intention for Control Plane
	// stacks on the infrastructure maintenance level is to run multiple ones if
	// desired or necessary. Reason for this may be scaling, management, or
	// refactoring of the underlying infrastructure of AWS resources.
	if len(list.Items) == 0 {
		return false, nil
	}

	return true, nil
}

func (r *Resource) g8sControlPlaneCRExists(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (bool, error) {
	o := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.Cluster, key.ClusterID(&cr)),
	}

	list, err := r.k8sClient.G8sClient().InfrastructureV1alpha2().G8sControlPlanes(cr.GetNamespace()).List(o)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// We simply list Control Plane CRs by filtering for the cluster ID label.
	// This may lead to one or more results as the intention for Control Plane
	// stacks on the infrastructure maintenance level is to run multiple ones if
	// desired or necessary. Reason for this may be scaling, management, or
	// refactoring of the underlying infrastructure of AWS resources.
	if len(list.Items) == 0 {
		return false, nil
	}

	return true, nil
}
