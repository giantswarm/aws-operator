package v31

import (
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "nodepools",
				Description: "Add testing version 6.6.0",
				Kind:        versionbundle.KindChanged,
			},
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
