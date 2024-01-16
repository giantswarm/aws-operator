package provider

import (
	"github.com/giantswarm/aws-operator/v16/flag/service/installation/guest/kubernetes/api/auth/provider/oidc"
)

type Provider struct {
	OIDC oidc.OIDC
}
