package v11

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Changed logging buckets to be deleted on test environments.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Added support for advanced monitoring in EC2.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Added Kubernetes API server whitelisting for NAT gateway EIPs.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Added support for disabling Route53.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Added support for making pause container configurable.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Added support for not including S3 bucket tags.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Allow port 2379 from host cluster to add support for etcd backup.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kubernetes",
				Description: "Updated to 1.10.3.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed kube-state-metrics so it can be managed by chart-operator.",
				Kind:        versionbundle.KindRemoved,
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
				Version: "1.10.3",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "aws-operator",
		Time:         time.Date(2018, time.May, 1, 11, 50, 0, 0, time.UTC),
		Version:      "3.1.1",
		WIP:          true,
	}
}
