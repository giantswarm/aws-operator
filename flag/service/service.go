package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/aws-operator/flag/service/aws"
	"github.com/giantswarm/aws-operator/flag/service/guest"
	"github.com/giantswarm/aws-operator/flag/service/installation"
)

type Service struct {
	AWS            aws.AWS
	Guest          guest.Guest
	Installation   installation.Installation
	Kubernetes     kubernetes.Kubernetes
	RegistryDomain string
}
