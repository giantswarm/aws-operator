package v29patch1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cloudformation",
				Description: "Duplicate etcd record set into public hosted zone.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudformation",
				Description: "Add ingress internal load-balancer in private hosted zone.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudformation",
				Description: "Use private subnets for internal Kubernetes API loadbalancer.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.8.2",
			},
			{
				Name:    "containerlinux",
				Version: "2135.4.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.13",
			},
			{
				Name:    "kubernetes",
				Version: "1.14.6",
			},
		},
		Name:    "aws-operator",
		Version: "5.3.1",
	}
}
