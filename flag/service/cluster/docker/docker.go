package docker

import (
	"github.com/giantswarm/aws-operator/v12/flag/service/cluster/docker/daemon"
)

type Docker struct {
	Daemon daemon.Daemon
}
