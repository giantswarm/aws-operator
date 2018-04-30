package cloudformation

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v10/ebs"
)

type EBSServiceMock struct {
}

func (e *EBSServiceMock) DeleteVolume(ctx context.Context, volumeID string) error {
	return nil
}

func (e *EBSServiceMock) DetachVolume(ctx context.Context, volumeID string, attachment ebs.VolumeAttachment, force bool, shutdown bool, wait bool) error {
	return nil
}

// ListVolumes always returns a list containing one volume because this is what
// the update process of the cloudformation resource needs.
func (e *EBSServiceMock) ListVolumes(customObject v1alpha1.AWSConfig, etcdVolume bool, persistentVolume bool) ([]ebs.Volume, error) {
	return []ebs.Volume{{}}, nil
}
