package cluster

import (
	"github.com/giantswarm/aws-operator/flag/service/cluster/calico"
	"github.com/giantswarm/aws-operator/flag/service/cluster/docker"
	"github.com/giantswarm/aws-operator/flag/service/cluster/dockerhub"
	"github.com/giantswarm/aws-operator/flag/service/cluster/kubernetes"
)

type Cluster struct {
	Calico     calico.Calico
	Docker     docker.Docker
	Dockerhub  dockerhub.Dockerhub
	Kubernetes kubernetes.Kubernetes
}
