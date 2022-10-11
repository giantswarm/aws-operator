package docker

import (
	"github.com/giantswarm/aws-operator/v14/flag/service/cluster/docker/daemon"
)

type Docker struct {
	Daemon daemon.Daemon
}
