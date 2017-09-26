package service

import (
	"github.com/giantswarm/aws-operator/flag/service/aws"
	"github.com/giantswarm/aws-operator/flag/service/kubernetes"
)

type Service struct {
	AWS        aws.AWS
	Kubernetes kubernetes.Kubernetes
}
