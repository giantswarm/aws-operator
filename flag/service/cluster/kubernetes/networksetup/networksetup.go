package networksetup

import (
	"github.com/giantswarm/aws-operator/v12/flag/service/cluster/kubernetes/networksetup/docker"
)

type NetworkSetup struct {
	Docker docker.Docker
}
