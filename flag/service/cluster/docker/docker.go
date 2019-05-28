package docker

import (
	"github.com/giantswarm/aws-operator/flag/service/cluster/docker/daemon"
)

type Docker struct {
	Daemon daemon.Daemon
}
