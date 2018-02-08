package kubernetes

import (
	"github.com/giantswarm/aws-operator/flag/service/installation/guest/kubernetes/api"
)

type Kubernetes struct {
	API api.API
}
