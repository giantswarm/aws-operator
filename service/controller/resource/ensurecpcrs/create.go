package ensurecpcrs

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

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

func (r *Resource) createAWSControlPlaneCR(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *Resource) createG8sControlPlaneCR(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *Resource) awsControlPlaneCRExists(ctx context.Context, obj interface{}) (bool, error) {
	return false, nil
}

func (r *Resource) g8sControlPlaneCRExists(ctx context.Context, obj interface{}) (bool, error) {
	return false, nil
}
