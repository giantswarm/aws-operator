package awsclient

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if errors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not yet availabile")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	err = r.addAWSClientsToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
