package adapter

import "github.com/giantswarm/aws-operator/service/controller/v15/key"

type hostIamRolesAdapter struct {
	PeerAccessRoleName string
	GuestAccountID     string
}

func (h *hostIamRolesAdapter) getHostIamRoles(cfg Config) error {
	h.PeerAccessRoleName = key.PeerAccessRoleName(cfg.CustomObject)
	h.GuestAccountID = cfg.GuestAccountID

	return nil
}
