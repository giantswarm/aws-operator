package versionbundle

import (
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/versionbundle"
)

func New() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cloudformation",
				Description: "Add IAMManager IAM role for kiam managed app.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudformation",
				Description: "Add Route53Manager IAM role for external-dns managed app.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "calico",
				Description: "Updated from v3.8.2 to v3.9.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "containerlinux",
				Description: "Updated from v2135.4.0 to v2191.5.0.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "etcd",
				Description: "Updated from v3.3.13 to v3.3.15.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Updated from v1.14.6 to v1.15.5.",
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
				Version: "18.06.1",
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
		Name:    project.Name(),
		Version: project.BundleVersion(),
	}
}

func NewSlice() []versionbundle.Bundle {
	return []versionbundle.Bundle{
		New(),
	}
}
