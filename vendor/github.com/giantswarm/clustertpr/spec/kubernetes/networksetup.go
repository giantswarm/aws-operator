package kubernetes

import "github.com/giantswarm/clustertpr/spec/kubernetes/networksetup"

type NetworkSetup struct {
	Docker networksetup.Docker `json:"docker" yaml:"docker"`
}
