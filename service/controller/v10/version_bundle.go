package v10

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cloudconfig",
				Description: "Enabled volume resizing feature.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Masked systemd-networkd-wait-online unit.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Fixed unencrypted encryption key injection via Cloud Config S3 uploads.",
				Kind:        versionbundle.KindSecurity,
			},
			{
				Component:   "aws-operator",
				Description: "Fixed idempotency of format-var-lib-docker service.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Detach EBS volumes before deletion when deleting clusters.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Added S3 access logs for cluster buckets.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Opened port 4194 for cAdvisor scraping from host cluster.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Fixed updating master nodes on all kinds of cluster updates.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "kubernetes",
				Description: "Updated to 1.10.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Added support for k8s API whitelisting.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "containerlinux",
				Description: "Updated to 1688.5.3.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated kube-state-metrics to version 1.3.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Changed kubelet bind mount mode from shared to rshared.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Disabled etcd3-defragmentation service in favor systemd timer.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Added /lib/modules mount for kubelet.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated CoreDNS to 1.1.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated Calico to 3.0.5.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated Etcd to 3.3.3.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed docker flag --disable-legacy-registry.",
				Kind:        versionbundle.KindRemoved,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed calico-ipip-pinger.",
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
				Version: "1.10.1",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
		},
		Name:    "aws-operator",
		Version: "3.1.0",
	}
}
