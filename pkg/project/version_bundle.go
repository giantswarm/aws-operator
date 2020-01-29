package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "TODO.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/X",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.10.1",
			},
			{
				Name:    "containerlinux",
				Version: "2247.6.0",
			},
			{
				Name:    "docker",
				Version: "18.06.3",
			},
			{
				Name:    "etcd",
				Version: "3.3.17",
			},
			{
				Name:    "kubernetes",
				Version: "1.16.3",
			},
		},
		Name:    Name(),
		Version: BundleVersion(),
	}
}
