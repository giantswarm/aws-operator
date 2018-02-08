package api

import (
	"github.com/giantswarm/aws-operator/flag/service/installation/guest/kubernetes/api/auth"
)

type API struct {
	Auth auth.Auth
}
