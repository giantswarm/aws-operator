package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Update to support Kubernetes 1.16.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2080",
				},
			},
			{
				Component:   "cloudconfig",
				Description: "Fix pause container image repository for China.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "vault",
				Description: "Fix vault encrypter role with new nodepools iam role names.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "kubernetes",
				Description: "Add Deny All as default Network Policy in kube-system and giantswarm namespaces.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/k8scloudconfig/pull/609",
				},
			},
			{
				Component:   "calico",
				Description: "Update from v3.9.1 to v3.10.1.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "containerlinux",
				Description: "Update from v2191.5.0 to v2247.6.0.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "etcd",
				Description: "Update from v3.3.15 to v3.3.17.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "kubernetes",
				Description: "Update from v1.15.5 to v1.16.3.",
				Kind:        versionbundle.KindAdded,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2084",
				},
			},
			{
				Component:   "cloudformation",
				Description: "Propagate tag name from ASG to EC2 instances.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2110",
				},
			},
			{
				Component:   "cloudformation",
				Description: "Drain nodes when deleting Node Pools.",
				Kind:        versionbundle.KindFixed,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2111",
				},
			},
			{
				Component:   "cloudformation",
				Description: "Encrypt blockdevice mappings in worker nodes.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/aws-operator/pull/2116",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.10.1",
			},
			{
				Name:    "containerlinux",
				Version: "2247.6.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.17",
			},
			{
				Name:    "kubernetes",
				Version: "1.16.3",
			},
		},
		Name:    Name(),
		Version: BundleVersion(),
	}
}
