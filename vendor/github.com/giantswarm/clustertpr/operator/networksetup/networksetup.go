package networksetup

import (
	"github.com/giantswarm/clustertpr/operator/networksetup/docker"
)

type NetworkSetup struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
