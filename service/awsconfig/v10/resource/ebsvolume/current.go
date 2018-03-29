package ebsvolume

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get both the Etcd volume and any Persistent Volumes.
	etcdVolume := true
	persistentVolume := true

	volumes, err := r.service.ListVolumes(customObject, etcdVolume, persistentVolume)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentState := &EBSVolumeState{
		Volumes: volumes,
	}

	return currentState, nil
}
