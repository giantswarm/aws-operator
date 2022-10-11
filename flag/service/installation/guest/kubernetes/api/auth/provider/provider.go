package provider

import (
	"github.com/giantswarm/aws-operator/v14/flag/service/installation/guest/kubernetes/api/auth/provider/oidc"
)

type Provider struct {
	OIDC oidc.OIDC
}
