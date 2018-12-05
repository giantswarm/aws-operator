package guest

import (
	"github.com/giantswarm/aws-operator/flag/service/guest/ignition"
	"github.com/giantswarm/aws-operator/flag/service/guest/ssh"
	"github.com/giantswarm/aws-operator/flag/service/guest/update"
)

type Guest struct {
	Ignition ignition.Ignition
	SSH      ssh.SSH
	Update   update.Update
}
