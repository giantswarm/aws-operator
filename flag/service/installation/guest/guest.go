package guest

import (
	"github.com/giantswarm/aws-operator/v13/flag/service/installation/guest/ipam"
	"github.com/giantswarm/aws-operator/v13/flag/service/installation/guest/kubernetes"
)

type Guest struct {
	IPAM       ipam.IPAM
	Kubernetes kubernetes.Kubernetes
}
