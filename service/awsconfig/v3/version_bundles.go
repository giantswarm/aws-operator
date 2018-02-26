package v3

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundles() []versionbundle.Bundle {
	return []versionbundle.Bundle{
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "cloudconfig",
					Description: "Fix eventual decryption race.",
					Kind:        versionbundle.KindFixed,
				},
			},
			Components: []versionbundle.Component{
				{
					Name:    "calico",
					Version: "3.0.1",
				},
				{
					Name:    "docker",
					Version: "17.09.0",
				},
				{
					Name:    "etcd",
					Version: "3.2.7",
				},
				{
					Name:    "coredns",
					Version: "1.0.5",
				},
				{
					Name:    "kubernetes",
					Version: "1.9.2",
				},
				{
					Name:    "nginx-ingress-controller",
					Version: "0.10.2",
				},
			},
			Dependencies: []versionbundle.Dependency{},
			Deprecated:   true,
			Name:         "aws-operator",
			Time:         time.Date(2018, time.January, 30, 11, 55, 0, 0, time.UTC),
			Version:      "2.0.1",
			WIP:          false,
		},
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "cloudformation",
					Description: "Add missing region to host account credentials.",
					Kind:        versionbundle.KindFixed,
				},
			},
			Components: []versionbundle.Component{
				{
					Name:    "calico",
					Version: "3.0.1",
				},
				{
					Name:    "docker",
					Version: "17.09.0",
				},
				{
					Name:    "etcd",
					Version: "3.2.7",
				},
				{
					Name:    "coredns",
					Version: "1.0.5",
				},
				{
					Name:    "kubernetes",
					Version: "1.9.2",
				},
				{
					Name:    "nginx-ingress-controller",
					Version: "0.10.2",
				},
			},
			Dependencies: []versionbundle.Dependency{},
			Deprecated:   true,
			Name:         "aws-operator",
			Time:         time.Date(2018, time.January, 31, 10, 43, 0, 0, time.UTC),
			Version:      "2.0.2",
			WIP:          false,
		},
	}
}
