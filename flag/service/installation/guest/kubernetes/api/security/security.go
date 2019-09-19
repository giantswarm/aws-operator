package security

import "github.com/giantswarm/aws-operator/flag/service/installation/guest/kubernetes/api/security/whitelist"

type Security struct {
	Whitelist whitelist.Whitelist
}
