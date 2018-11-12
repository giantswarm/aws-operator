package v18patch1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cloudconfig",
				Description: "Updated k8scloudconfig to 3.7.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kube-proxy",
				Description: "Now gets installed and upgraded before Calico.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "calico",
				Description: "Reapplied missing priority class.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.2.3",
			},
			{
				Name:    "containerlinux",
				Version: "1855.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.9",
			},
			{
				Name:    "kubernetes",
				Version: "1.12.2",
			},
		},
		Name:    "aws-operator",
		Version: "4.3.1",
	}
}
