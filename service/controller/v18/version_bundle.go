package v18

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Update Kubernetes to 1.12.1. More info here: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG-1.12.md",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "calico",
				Description: "Updated Calico manifest with resource limits to get QoS policy guaranteed.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Enabled admission plugins: DefaultTolerationSeconds, MutatingAdmissionWebhook, ValidatingAdmissionWebhook.",
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
		Version: "4.3.0",
	}
}
