package v5

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Delete AWS Cloud Provider resources when deleting clusters.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Add AWS resource tag with GiantSwarm Cluster ID for reporting.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Change default etcd data dir to /var/lib/etcd.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Allow persistent volumes to be automatically extended when claims are changed.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Fixed problem with deleting the Etcd EBS volume.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "Calico",
				Description: "Updated to 3.0.2.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "Etcd",
				Description: "Updated to 3.3.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "CoreDNS",
				Description: "Updated to 1.0.6.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "Nginx Ingress Controller",
				Description: "Updated to 0.11.0.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Tune Kubelet flags for protecting key units (Kubelet and Container Runtime) from workload overloads.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.0.2",
			},
			{
				Name:    "containerlinux",
				Version: "1576.5.0",
			},
			{
				Name:    "docker",
				Version: "17.09.0",
			},
			{
				Name:    "etcd",
				Version: "3.3.1",
			},
			{
				Name:    "coredns",
				Version: "1.0.6",
			},
			{
				Name:    "kubernetes",
				Version: "1.9.2",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.11.0",
			},
		},
		Name:    "aws-operator",
		Version: "2.1.1",
	}
}
