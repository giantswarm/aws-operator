package networksetup

import (
	"github.com/giantswarm/aws-operator/v16/flag/service/cluster/kubernetes/networksetup/docker"
)

type NetworkSetup struct {
	Docker docker.Docker
}
