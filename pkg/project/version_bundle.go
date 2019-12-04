package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cloudformation",
				Description: "Bring back name tags to AWS resources like VPCs, Subnets and EC2 Instances.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2059",
				},
			},
			{
				Component:   "kubelet",
				Description: "Label nodes with operator version instead of release version.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2064",
				},
			},
			{
				Component:   "calico",
				Description: "Updated from v3.9.1 to v3.10.1.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "containerlinux",
				Description: "Updated from v2191.5.0 to v2247.6.0.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "etcd",
				Description: "Updated from v3.3.15 to v3.3.17.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "kubernetes",
				Description: "Updated from v1.15.5 to v1.16.3.",
				Kind:        versionbundle.KindAdded,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
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
