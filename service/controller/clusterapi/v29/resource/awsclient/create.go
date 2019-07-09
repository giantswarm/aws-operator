package awsclient

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addAWSClientsToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
