package installation

import (
	"github.com/giantswarm/aws-operator/v16/flag/service/installation/guest"
)

type Installation struct {
	Name  string
	Guest guest.Guest
}
