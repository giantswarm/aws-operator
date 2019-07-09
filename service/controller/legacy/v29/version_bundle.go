package v29

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "ignition",
				Description: "Add name label for default and kube-system namespaces.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.7.2",
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
				Version: "1.14.3",
			},
		},
		Name:    "aws-operator",
		Version: "5.3.0",
	}
}
