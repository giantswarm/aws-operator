package v11

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
<<<<<<< HEAD
				Component:   "aws-operator",
				Description: "Changed logging buckets to be deleted on test environments.",
=======
				Component:   "component",
				Description: "Put your description here.",
>>>>>>> master
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.0.5",
			},
			{
				Name:    "containerlinux",
				Version: "1688.5.3",
			},
			{
				Name:    "docker",
				Version: "17.12.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.3",
			},
			{
				Name:    "coredns",
				Version: "1.1.1",
			},
			{
				Name:    "kubernetes",
				Version: "1.10.1",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "aws-operator",
<<<<<<< HEAD
		Time:         time.Date(2018, time.May, 1, 11, 50, 0, 0, time.UTC),
=======
		Time:         time.Date(2018, time.April, 30, 18, 50, 0, 0, time.UTC),
>>>>>>> master
		Version:      "3.1.1",
		WIP:          true,
	}
}
