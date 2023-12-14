package service

import (
	"github.com/giantswarm/operatorkit/v8/pkg/flag/service/kubernetes"

	"github.com/giantswarm/aws-operator/v15/flag/service/aws"
	"github.com/giantswarm/aws-operator/v15/flag/service/cluster"
	"github.com/giantswarm/aws-operator/v15/flag/service/guest"
	"github.com/giantswarm/aws-operator/v15/flag/service/installation"
	"github.com/giantswarm/aws-operator/v15/flag/service/registry"
)

type Service struct {
	AWS          aws.AWS
	Cluster      cluster.Cluster
	Guest        guest.Guest
	Installation installation.Installation
	Kubernetes   kubernetes.Kubernetes
	Registry     registry.Registry
}
