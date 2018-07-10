package v13

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "aws-operator",
				Description: "Updated EC2 and S3 endpoints for supporting AWS China regions.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Added CA public key into trusted user keys for SSO ssh.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Made master docker volume ephemeral.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "aws-operator",
				Description: "Fixed volume attachments while upgrading.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "aws-operator",
				Description: "Added Vault encrypter as an alternative to KMS (especially for China where it is disabled).",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Enabled migration resource with credentials logic.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Add public routes for accessing Vault ELB when using Vault encrypter.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Made docker registry domain for component images configurable.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "aws-operator",
				Description: "Set version bundle version stack output via parameter.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.0.5",
			},
			{
				Name:    "containerlinux",
				Version: "1745.4.0",
			},
			{
				Name:    "docker",
				Version: "18.03.1",
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
				Version: "1.10.4",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
		},
		Name:    "aws-operator",
		Version: "3.2.0",
	}
}
