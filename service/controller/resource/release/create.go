package release

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addReleaseToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
