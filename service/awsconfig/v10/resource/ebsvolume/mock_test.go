package ebsvolume

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/ebs"
)

type EBSServiceMock struct {
	volumes []ebs.Volume
}

func (e *EBSServiceMock) DeleteVolume(ctx context.Context, volumeID string) error {
	return nil
}

func (e *EBSServiceMock) DetachVolume(ctx context.Context, volumeID string, attachment ebs.VolumeAttachment, force bool, shutdown bool) error {
	return nil
}

func (e *EBSServiceMock) ListVolumes(customObject v1alpha1.AWSConfig, etcdVolume bool, persistentVolume bool) ([]ebs.Volume, error) {
	return e.volumes, nil
}
