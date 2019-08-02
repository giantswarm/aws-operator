package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/aws-operator/flag/service/aws"
	"github.com/giantswarm/aws-operator/flag/service/cluster"
	"github.com/giantswarm/aws-operator/flag/service/feature"
	"github.com/giantswarm/aws-operator/flag/service/guest"
	"github.com/giantswarm/aws-operator/flag/service/installation"
	"github.com/giantswarm/aws-operator/flag/service/test"
)

type Service struct {
	AWS            aws.AWS
	Cluster        cluster.Cluster
	Feature        feature.Feature
	Guest          guest.Guest
	Installation   installation.Installation
	Kubernetes     kubernetes.Kubernetes
	RegistryDomain string
	Test           test.Test
}
