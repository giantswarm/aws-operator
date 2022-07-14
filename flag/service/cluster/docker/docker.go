package docker

import (
	"github.com/giantswarm/aws-operator/v2/flag/service/cluster/docker/daemon"
)

type Docker struct {
	Daemon daemon.Daemon
}
