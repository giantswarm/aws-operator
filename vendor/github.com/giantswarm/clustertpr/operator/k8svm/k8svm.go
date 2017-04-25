package k8svm

import (
	"github.com/giantswarm/clustertpr/operator/k8svm/docker"
)

type K8sVM struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
