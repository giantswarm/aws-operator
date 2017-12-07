package kmskeyv1

import (
	"context"

	"github.com/giantswarm/aws-operator/service/keyv1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	desiredState := KMSKeyState{}

	customObject, err := keyv1.ToCustomObject(obj)
	if err != nil {
		return desiredState, err
	}

	clusterID := keyv1.ClusterID(customObject)
	desiredState.KeyAlias = toAlias(clusterID)

	return desiredState, nil
}
