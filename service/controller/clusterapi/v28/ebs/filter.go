package ebs

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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

func NewDockerVolumeFilter(cr v1alpha1.Cluster) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == nameTagKey && *t.Value == key.VolumeNameDocker(cr) {
			return true
		}
		return false
	}
}

func NewEtcdVolumeFilter(cr v1alpha1.Cluster) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == nameTagKey && *t.Value == key.VolumeNameEtcd(cr) {
			return true
		}
		return false
	}
}

func NewPersistentVolumeFilter(cr v1alpha1.Cluster) func(t *ec2.Tag) bool {
	return func(t *ec2.Tag) bool {
		if *t.Key == cloudProviderPersistentVolumeTagKey {
			return true
		}
		return false
	}
}
