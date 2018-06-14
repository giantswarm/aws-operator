package ebs

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v14/key"
)

const (
	cloudProviderPersistentVolumeTagKey = "kubernetes.io/created-for/pv/name"
	nameTagKey                          = "Name"
)

func IsFiltered(vol *ec2.Volume, filterFuncs []func(t *ec2.Tag) bool) bool {
	for _, f := range filterFuncs {
		for _, t := range vol.Tags {
			if f(t) {
				return true
			}
		}
	}

	return false
}

func NewDockerVolumeFilter(customObject v1alpha1.AWSConfig) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == nameTagKey && *t.Value == key.DockerVolumeName(customObject) {
			return true
		}
		return false
	}
}

func NewEtcdVolumeFilter(customObject v1alpha1.AWSConfig) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == nameTagKey && *t.Value == key.EtcdVolumeName(customObject) {
			return true
		}
		return false
	}
}

func NewPersistentVolumeFilter(customObject v1alpha1.AWSConfig) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == cloudProviderPersistentVolumeTagKey {
			return true
		}
		return false
	}
}
