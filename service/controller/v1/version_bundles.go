package v1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundles() []versionbundle.Bundle {
	return []versionbundle.Bundle{
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "calico",
					Description: "Calico version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "docker",
					Description: "Docker version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "etcd",
					Description: "Etcd version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "kubedns",
					Description: "KubeDNS version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "kubernetes",
					Description: "Kubernetes version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "nginx-ingress-controller",
					Description: "Nginx-ingress-controller version updated.",
					Kind:        versionbundle.KindChanged,
				},
			},
			Components: []versionbundle.Component{
				{
					Name:    "calico",
					Version: "2.6.2",
				},
				{
					Name:    "docker",
					Version: "1.12.6",
				},
				{
					Name:    "etcd",
					Version: "3.2.7",
				},
				{
					Name:    "kubedns",
					Version: "1.14.5",
				},
				{
					Name:    "kubernetes",
					Version: "1.8.1",
				},
				{
					Name:    "nginx-ingress-controller",
					Version: "0.9.0",
				},
			},
			Name:    "aws-operator",
			Version: "0.1.0",
		},
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "kubernetes",
					Description: "Updated to kubernetes 1.8.4. Fixes a goroutine leak in the k8s api.",
					Kind:        versionbundle.KindChanged,
				},
			},
			Components: []versionbundle.Component{
				{
					Name:    "calico",
					Version: "2.6.2",
				},
				{
					Name:    "docker",
					Version: "1.12.6",
				},
				{
					Name:    "etcd",
					Version: "3.2.7",
				},
				{
					Name:    "kubedns",
					Version: "1.14.5",
				},
				{
					Name:    "kubernetes",
					Version: "1.8.4",
				},
				{
					Name:    "nginx-ingress-controller",
					Version: "0.9.0",
				},
			},
			Name:    "aws-operator",
			Version: "1.0.0",
		},
	}
}
