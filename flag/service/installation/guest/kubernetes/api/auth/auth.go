package auth

import (
	"github.com/giantswarm/aws-operator/v14/flag/service/installation/guest/kubernetes/api/auth/provider"
)

type Auth struct {
	Provider provider.Provider
}
