package service

import (
	"github.com/giantswarm/kvm-operator/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
