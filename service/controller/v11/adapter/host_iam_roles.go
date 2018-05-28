package adapter

import "github.com/giantswarm/aws-operator/service/controller/v10/key"

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v10/templates/cloudformation/hostpre/iam_roles.go
//

type hostIamRolesAdapter struct {
	PeerAccessRoleName string
	GuestAccountID     string
}

func (h *hostIamRolesAdapter) getHostIamRoles(cfg Config) error {
	h.PeerAccessRoleName = key.PeerAccessRoleName(cfg.CustomObject)
	h.GuestAccountID = cfg.GuestAccountID

	return nil
}
