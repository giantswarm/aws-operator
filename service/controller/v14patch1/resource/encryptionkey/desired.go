package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredState, err := r.encrypter.DesiredState(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredState, nil
}
