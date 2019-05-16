package kubernetes

import (
	"github.com/giantswarm/aws-operator/flag/service/cluster/kubernetes/api"
	"github.com/giantswarm/aws-operator/flag/service/cluster/kubernetes/networksetup"
	"github.com/giantswarm/aws-operator/flag/service/cluster/kubernetes/ssh"
)

type Kubernetes struct {
	API          api.API
	NetworkSetup networksetup.NetworkSetup
	SSH          ssh.SSH
}
