package adapter

import "github.com/giantswarm/aws-operator/service/controller/v17patch2/key"

type HostPreIAMRolesAdapter struct {
	PeerAccessRoleName string
	GuestAccountID     string
}

func (h *HostPreIAMRolesAdapter) Adapt(cfg Config) error {
	h.PeerAccessRoleName = key.PeerAccessRoleName(cfg.CustomObject)
	h.GuestAccountID = cfg.GuestAccountID

	return nil
}
