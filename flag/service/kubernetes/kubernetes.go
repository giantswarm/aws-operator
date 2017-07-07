package kubernetes

import (
	"github.com/giantswarm/aws-operator/flag/service/kubernetes/tls"
)

type Kubernetes struct {
	Address     string
	InCluster   string
	Insecure    string
	Password    string
	TLS         tls.TLS
	BearerToken string
	Username    string
}
