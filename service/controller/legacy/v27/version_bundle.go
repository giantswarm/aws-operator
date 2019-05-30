package v27

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Ensure the usage of the correct encryption key for AWS China.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "aws-operator",
				Description: "Prevent race conditions in decryption units for AWS China.",
				Kind:        versionbundle.KindFixed,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.6.1",
			},
			{
				Name:    "containerlinux",
				Version: "2023.5.0",
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
				Version: "1.14.1",
			},
		},
		Name:    "aws-operator",
		Version: "5.1.0",
	}
}
