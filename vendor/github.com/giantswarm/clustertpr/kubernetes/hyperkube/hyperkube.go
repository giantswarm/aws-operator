package hyperkube

import (
	"github.com/giantswarm/clustertpr/kubernetes/hyperkube/docker"
)

type Hyperkube struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
