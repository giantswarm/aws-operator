package aws

import (
	"github.com/giantswarm/aws-operator/flag/service/aws/guest"
	"github.com/giantswarm/aws-operator/flag/service/aws/host"
)

type AWS struct {
	Guest guest.Guest
	Host  host.Host
}
