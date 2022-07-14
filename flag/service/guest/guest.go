package guest

import (
	"github.com/giantswarm/aws-operator/v12/flag/service/guest/ignition"
	"github.com/giantswarm/aws-operator/v12/flag/service/guest/ssh"
)

type Guest struct {
	Ignition ignition.Ignition
	SSH      ssh.SSH
}
