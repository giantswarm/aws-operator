package v23

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Update kubernetes to 1.13.3",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "etcd",
				Description: "Update etcd to 3.3.12",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "calico",
				Description: "Update calico to 3.5.1",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.5.1",
			},
			{
				Name:    "containerlinux",
				Version: "1967.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.12",
			},
			{
				Name:    "kubernetes",
				Version: "1.13.3",
			},
		},
		Name:    "aws-operator",
		Version: "4.7.0",
	}
}
