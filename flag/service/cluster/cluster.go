package cluster

import (
	"github.com/giantswarm/aws-operator/v12/flag/service/cluster/calico"
	"github.com/giantswarm/aws-operator/v12/flag/service/cluster/docker"
	"github.com/giantswarm/aws-operator/v12/flag/service/cluster/kubernetes"
)

type Cluster struct {
	Calico     calico.Calico
	Docker     docker.Docker
	Kubernetes kubernetes.Kubernetes
}
