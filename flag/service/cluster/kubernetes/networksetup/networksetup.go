package networksetup

import (
	"github.com/giantswarm/aws-operator/v15/flag/service/cluster/kubernetes/networksetup/docker"
)

type NetworkSetup struct {
	Docker docker.Docker
}
