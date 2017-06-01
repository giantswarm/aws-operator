package kubectl

import "github.com/giantswarm/clustertpr/kubernetes/kubectl/docker"

type Kubectl struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
