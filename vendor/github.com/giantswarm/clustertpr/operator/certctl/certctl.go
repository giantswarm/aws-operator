package certctl

import (
	"github.com/giantswarm/clustertpr/operator/certctl/docker"
)

type Certctl struct {
	Docker docker.Docker
}
