package kubernetes

import "github.com/giantswarm/clustertpr/spec/kubernetes/kubectl"

type Kubectl struct {
	Docker kubectl.Docker `json:"docker" yaml:"docker"`
}
