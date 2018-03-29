package ebsvolume

import (
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/ebs"
)

type EBSVolumeState struct {
	Volumes []ebs.Volume
}
