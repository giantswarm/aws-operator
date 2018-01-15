package adapter

import "github.com/giantswarm/aws-operator/service/keyv2"

// template related to this adapter: service/templates/cloudformation/host-pre/iam_roles.yaml

type hostIamRolesAdapter struct {
	PeerAccessRoleName string
	GuestAccountID     string
}

func (h *hostIamRolesAdapter) getHostIamRoles(cfg Config) error {
	h.PeerAccessRoleName = keyv2.PeerAccessRoleName(cfg.CustomObject)
	h.GuestAccountID = cfg.GuestAccountID

	return nil
}
