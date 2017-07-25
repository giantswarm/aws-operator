package kubernetes

import "github.com/giantswarm/clustertpr/spec/kubernetes/hyperkube"

type Hyperkube struct {
	Docker hyperkube.Docker `json:"docker" yaml:"docker"`
}
