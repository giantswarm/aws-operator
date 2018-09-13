package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch2/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentState, err := r.encrypter.CurrentState(ctx, customObject)
	if err != nil {
		return currentState, microerror.Mask(err)
	}

	return currentState, nil
}
