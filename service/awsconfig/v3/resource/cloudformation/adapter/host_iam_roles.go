package adapter

import "github.com/giantswarm/aws-operator/service/awsconfig/v3/key"

// template related to this adapter: service/templates/cloudformation/host-pre/iam_roles.yaml

type hostIamRolesAdapter struct {
	PeerAccessRoleName string
	GuestAccountID     string
}

func (h *hostIamRolesAdapter) getHostIamRoles(cfg Config) error {
	h.PeerAccessRoleName = key.PeerAccessRoleName(cfg.CustomObject)
	h.GuestAccountID = cfg.GuestAccountID

	return nil
}
