package key

import (
	"fmt"

	"github.com/giantswarm/aws-operator/integration/env"
)

func HostPeerStackName() string {
	return fmt.Sprintf("host-peer-%s", env.ClusterID())
}
