package v24

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Mount /var/log directory in an EBS Volume.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "containerlinux",
				Description: "Update CoreOS to 2023.4.0.(CVE-2019-8912).",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Use proper hostname annotation for nodes.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Update to 1.13.4 (CVE-2019-1002100).",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Fix update process node termination.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.5.1",
			},
			{
				Name:    "containerlinux",
				Version: "2023.4.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.12",
			},
			{
				Name:    "kubernetes",
				Version: "1.13.4",
			},
		},
		Name:    "aws-operator",
		Version: "4.8.0",
	}
}
