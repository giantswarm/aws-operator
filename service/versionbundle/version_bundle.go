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
				Component:   "clusterapi",
				Description: "Add cleanuprecordsets resource to cleanup non-managed route53 records.",
				Kind:        versionbundle.KindAdded,
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
		Name:    project.Name(),
		Version: project.BundleVersion(),
	}
}

func NewSlice() []versionbundle.Bundle {
	return []versionbundle.Bundle{
		New(),
	}
}
