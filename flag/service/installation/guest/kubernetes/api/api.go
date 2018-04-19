package api

import (
	"github.com/giantswarm/aws-operator/flag/service/installation/guest/kubernetes/api/auth"
	"github.com/giantswarm/aws-operator/flag/service/installation/guest/kubernetes/api/security"
)

type API struct {
	Auth     auth.Auth
	Security security.Security
}
