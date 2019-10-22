package v31

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "component",
				Description: "TODO",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.9.1",
			},
			{
				Name:    "containerlinux",
				Version: "2191.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.3",
			},
			{
				Name:    "etcd",
				Version: "3.3.15",
			},
			{
				Name:    "kubernetes",
				Version: "1.15.5",
			},
		},
		Name:    "aws-operator",
		Version: "5.5.0",
	}
}
