package service

import (
	"github.com/giantswarm/aws-operator/flag/service/aws"
	"github.com/giantswarm/aws-operator/flag/service/installation"
	"github.com/giantswarm/aws-operator/flag/service/kubernetes"
)

type Service struct {
	AWS          aws.AWS
	Installation installation.Installation
	Kubernetes   kubernetes.Kubernetes
}
