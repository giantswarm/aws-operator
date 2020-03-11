package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Modified to get component versions from releases",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "0.0.0",
			},
			{
				Name:    "containerlinux",
				Version: "0.0.0",
			},
			{
				Name:    "docker",
				Version: "0.0.0",
			},
			{
				Name:    "etcd",
				Version: "0.0.0",
			},
			{
				Name:    "kubernetes",
				Version: "0.0.0",
			},
		},
		Name:    Name(),
		Version: BundleVersion(),
	}
}
