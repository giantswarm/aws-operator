package v17patch2

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "containerlinux",
				Description: "Fix for CVE-2019-5736.",
				Kind:        versionbundle.KindSecurity,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.2.3",
			},
			{
				Name:    "containerlinux",
				Version: "1967.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.8",
			},
			{
				Name:    "kubernetes",
				Version: "1.11.5",
			},
		},
		Name:    "aws-operator",
		Version: "4.2.2",
	}
}
