package guest

import (
	"github.com/giantswarm/aws-operator/flag/service/installation/guest/kubernetes"
)

type Guest struct {
	Kubernetes kubernetes.Kubernetes
}
