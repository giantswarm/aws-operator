package v20

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Add support for using multiple availability zones. See https://docs.giantswarm.io/basics/multiaz/.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Switch back to an internal elb for etcd. Calico connection handling problems have been fixed upstream.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Upgrade: Terminate the old master right after detaching of its volumes.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Add autoscaling permissions to the IAM policy of the cluster.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloud-config",
				Description: "The pod priority class for calico got lost. We found it again!",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "k8s-addons",
				Description: "kube-proxy is now installed before calico during cluster creation and upgrades.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloud-config",
				Description: "Calico upgrade improvements: Remove the old master from the k8s api and wait until etcd DNS is resolvable before upgrading calico. Networking pods crashlooping isn't fun!",
				Kind:        versionbundle.KindAdded,
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
		Version: "4.4.0",
	}
}
