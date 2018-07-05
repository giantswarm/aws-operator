package service

import (
	"github.com/giantswarm/aws-operator/flag/service/aws"
	"github.com/giantswarm/aws-operator/flag/service/guest"
	"github.com/giantswarm/aws-operator/flag/service/installation"
	"github.com/giantswarm/aws-operator/flag/service/kubernetes"
)

type Service struct {
	AWS            aws.AWS
	Guest          guest.Guest
	Installation   installation.Installation
	Kubernetes     kubernetes.Kubernetes
	RegistryDomain string
}
