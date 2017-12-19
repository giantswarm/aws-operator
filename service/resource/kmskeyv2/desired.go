package kmskeyv2

import (
	"context"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	desiredState := KMSKeyState{}

	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return desiredState, err
	}

	clusterID := keyv2.ClusterID(customObject)
	desiredState.KeyAlias = toAlias(clusterID)

	return desiredState, nil
}
