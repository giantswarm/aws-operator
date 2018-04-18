package kmskey

import (
	"context"

	"github.com/giantswarm/aws-operator/service/controller/v5/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	desiredState := KMSKeyState{}

	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return desiredState, err
	}

	clusterID := key.ClusterID(customObject)
	desiredState.KeyAlias = toAlias(clusterID)

	return desiredState, nil
}
