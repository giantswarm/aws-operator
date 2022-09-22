package kubernetes

import (
	"github.com/giantswarm/aws-operator/v13/flag/service/cluster/kubernetes/api"
	"github.com/giantswarm/aws-operator/v13/flag/service/cluster/kubernetes/networksetup"
	"github.com/giantswarm/aws-operator/v13/flag/service/cluster/kubernetes/ssh"
)

type Kubernetes struct {
	API           api.API
	ClusterDomain string
	NetworkSetup  networksetup.NetworkSetup
	SSH           ssh.SSH
}
