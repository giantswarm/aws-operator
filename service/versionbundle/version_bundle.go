package versionbundle

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/aws-operator/pkg/project"
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
			{
				Component:   "nodepools",
				Description: "Add Node Pools functionality. See https://docs.giantswarm.io/basics/nodepools/ for details.",
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
