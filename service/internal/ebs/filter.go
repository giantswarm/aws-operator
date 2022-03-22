package ebs

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"

	"github.com/giantswarm/aws-operator/service/controller/key"
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

func NewDockerVolumeFilter(cr infrastructurev1alpha3.AWSCluster) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == nameTagKey && *t.Value == key.VolumeNameDocker(cr) {
			return true
		}
		return false
	}
}

func NewEtcdVolumeFilter(cr infrastructurev1alpha3.AWSCluster) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == nameTagKey && *t.Value == key.VolumeNameEtcd(cr) {
			return true
		}
		return false
	}
}

func NewPersistentVolumeFilter(cr infrastructurev1alpha3.AWSCluster) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		return *t.Key == cloudProviderPersistentVolumeTagKey
	}
}
