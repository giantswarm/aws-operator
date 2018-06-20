package guest

import "github.com/giantswarm/aws-operator/flag/service/guest/update"
import "github.com/giantswarm/aws-operator/flag/service/guest/ssh"

type Guest struct {
	Update update.Update
	SSH    ssh.SSH
}
