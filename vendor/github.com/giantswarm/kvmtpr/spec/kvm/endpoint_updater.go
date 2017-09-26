package kvm

import "github.com/giantswarm/kvmtpr/spec/kvm/endpointupdater"

type EndpointUpdater struct {
	Docker endpointupdater.Docker `json:"docker" yaml:"docker"`
}
