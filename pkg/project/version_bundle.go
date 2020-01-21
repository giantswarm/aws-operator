package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Update to support Kubernetes 1.16.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2090",
				},
			},
			{
				Component:   "calico",
				Description: "Update from v3.9.1 to v3.10.1.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/k8scloudconfig/pull/615",
				},
			},
			{
				Component:   "containerlinux",
				Description: "Update from v2191.5.0 to 2247.6.0.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/k8scloudconfig/pull/615",
				},
			},
			{
				Component:   "etcd",
				Description: "Update from v3.3.15 to v3.3.17.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/k8scloudconfig/pull/615",
				},
			},
			{
				Component:   "kubernetes",
				Description: "Update from v1.15.5 to v1.16.3.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/k8scloudconfig/pull/615",
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
