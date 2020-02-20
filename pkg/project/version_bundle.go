package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Fix AWS resource tags.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2137",
				},
			},
			{
				Component:   "aws-operator",
				Description: "Allow network traffic between Node Pools.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2141",
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
				Version: "18.06.1",
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
