package adapter

import "github.com/giantswarm/aws-operator/service/controller/v14/key"

type hostPreIAMRolesAdapter struct {
	PeerAccessRoleName string
	GuestAccountID     string
}

func (h *hostPreIAMRolesAdapter) Adapt(cfg Config) error {
	h.PeerAccessRoleName = key.PeerAccessRoleName(cfg.CustomObject)
	h.GuestAccountID = cfg.GuestAccountID

	return nil
}
